package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"pornboss/internal/common/logging"
	"pornboss/internal/util"

	"golang.org/x/net/html"
)

// CoverManager coordinates background cover downloads.
type CoverManager struct {
	tasks    chan string
	coverDir string
	workers  int
}

// NewCoverManager creates a manager when coverDir is provided. Nil is returned when empty.
func NewCoverManager(coverDir string) *CoverManager {
	if strings.TrimSpace(coverDir) == "" {
		return nil
	}
	return &CoverManager{
		tasks:    make(chan string, 5000), // larger buffer to reduce producer blocking
		coverDir: coverDir,
		workers:  10,
	}
}

// Start launches the worker; safe to call with nil manager.
func (m *CoverManager) Start(ctx context.Context) {
	if m == nil {
		return
	}
	if m.workers <= 0 {
		m.workers = 1
	}
	for i := 0; i < m.workers; i++ {
		go m.worker(ctx)
	}
}

// Enqueue schedules a cover download; blocks when queue is full.
func (m *CoverManager) Enqueue(code string) {
	if m == nil {
		return
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return
	}
	m.tasks <- code
}

// Exists reports whether a cover file already exists for the code (any known extension).
func (m *CoverManager) Exists(code string) bool {
	if m == nil {
		return false
	}
	_, ok := FindCoverPath(m.coverDir, code)
	return ok
}

func (m *CoverManager) worker(ctx context.Context) {
	if m == nil {
		return
	}
	_ = os.MkdirAll(m.coverDir, 0o755)
	for {
		select {
		case <-ctx.Done():
			return
		case code := <-m.tasks:
			if err := m.handleTask(ctx, code); err != nil {
				logging.Error("jav cover: code=%s err=%v", code, err)
			}
		}
	}
}

func (m *CoverManager) handleTask(parent context.Context, code string) error {
	code = normalizeCode(code)
	if code == "" {
		return errors.New("empty code")
	}
	if m.Exists(code) {
		return nil
	}

	ctx, cancel := context.WithTimeout(parent, 45*time.Second)
	defer cancel()

	coverURL, err := FetchCoverURL(ctx, code)
	if err != nil {
		if errors.Is(err, util.ErrCachedNotFound) {
			return nil
		}
		return err
	}
	if coverURL == "" {
		return errors.New("cover url not found")
	}

	if err := m.downloadCover(ctx, code, coverURL); err != nil {
		if errors.Is(err, util.ErrCachedNotFound) {
			return nil
		}
		return err
	}
	return nil
}

// FetchCoverURL retrieves the cover image URL from javdatabase for the given code.
func FetchCoverURL(ctx context.Context, code string) (string, error) {
	code = normalizeCode(code)
	if code == "" {
		return "", errors.New("empty code")
	}

	pageURL := fmt.Sprintf("https://www.javdatabase.com/movies/%s", url.PathEscape(code))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; JavCoverBot/1.0)")

	resp, err := util.DoRequest(req)
	if err != nil {
		if errors.Is(err, util.ErrCachedNotFound) {
			return "", err
		}
		return "", fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", util.ErrCachedNotFound
		}
		return "", fmt.Errorf("fetch page: status %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", fmt.Errorf("read page: %w", err)
	}

	coverURL, err := extractCoverURL(body)
	if err != nil {
		return "", fmt.Errorf("parse cover url: %w", err)
	}
	return strings.TrimSpace(coverURL), nil
}

func (m *CoverManager) downloadCover(ctx context.Context, code, coverURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, coverURL, nil)
	if err != nil {
		return fmt.Errorf("build cover request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; JavCoverBot/1.0)")
	resp, err := util.DoRequest(req)
	if err != nil {
		if errors.Is(err, util.ErrCachedNotFound) {
			return err
		}
		return fmt.Errorf("download cover: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return util.ErrCachedNotFound
		}
		return fmt.Errorf("download cover: status %s", resp.Status)
	}

	ext := strings.ToLower(path.Ext(resp.Request.URL.Path))
	if ext == "" || len(ext) > 5 {
		ext = guessExt(resp.Header.Get("Content-Type"))
	}
	if ext == "" {
		ext = ".jpg"
	}

	target := filepath.Join(m.coverDir, code+ext)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("ensure cover dir: %w", err)
	}
	tmp := target + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("write cover: %w", err)
	}
	out.Close()

	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("finalize cover: %w", err)
	}
	return nil
}

var knownExts = []string{".jpg", ".jpeg", ".png", ".webp"}

func normalizeCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}

// FindCoverPath returns the existing cover file path for the given code within dir.
func FindCoverPath(dir, code string) (string, bool) {
	code = normalizeCode(code)
	if code == "" {
		return "", false
	}
	for _, ext := range knownExts {
		p := filepath.Join(dir, code+ext)
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return "", false
}

func guessExt(ct string) string {
	ct = strings.ToLower(strings.TrimSpace(ct))
	switch {
	case strings.Contains(ct, "webp"):
		return ".webp"
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "jpeg"), strings.Contains(ct, "jpg"):
		return ".jpg"
	default:
		return ""
	}
}

func extractCoverURL(body []byte) (string, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("parse html: %w", err)
	}
	var cover string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if cover != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "meta" {
			var prop, content string
			for _, a := range n.Attr {
				switch strings.ToLower(a.Key) {
				case "property":
					prop = strings.ToLower(a.Val)
				case "content":
					content = a.Val
				}
			}
			if prop == "og:image" && strings.TrimSpace(content) != "" {
				cover = content
				return
			}
		}
		if n.Type == html.ElementNode && n.Data == "img" {
			if coverHasClass(n, "poster") || coverHasClass(n, "cover") {
				if src := coverAttr(n, "src"); src != "" {
					cover = src
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return strings.TrimSpace(cover), nil
}

func coverHasClass(n *html.Node, target string) bool {
	for _, a := range n.Attr {
		if a.Key != "class" {
			continue
		}
		parts := strings.Fields(a.Val)
		for _, p := range parts {
			if p == target {
				return true
			}
		}
	}
	return false
}

func coverAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}
