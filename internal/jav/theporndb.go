package jav

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pornboss/internal/common/logging"
	"pornboss/internal/util"
)

// ThePornDB implements JavLookupProvider.
type ThePornDB struct{}

var ThePornDBProvider JavLookupProvider = ThePornDB{}

const thePornDBBearerToken = "uqtWi1LRXC2ngClxz8QrqfOERuH2qbuh89CQAiXx85088612"

// LookupActressByCode implements JavLookupProvider.
func (ThePornDB) LookupActressByCode(code string) (*ActressInfo, error) {
	return nil, errors.New("theporndb: lookup actress not supported")
}

// LookupActressByJapaneseName implements JavLookupProvider.
func (ThePornDB) LookupActressByJapaneseName(name string) (*ActressInfo, error) {
	return nil, errors.New("theporndb: lookup actress not supported")
}

// LookupJavByCode implements JavLookupProvider.
func (ThePornDB) LookupJavByCode(code string) (*Info, error) {
	return nil, errors.New("theporndb: lookup jav not supported")
}

// LookupCoverURLByCode implements JavLookupProvider.
func (ThePornDB) LookupCoverURLByCode(code string) (string, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return "", ResourceNotFonud
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	targetURL := fmt.Sprintf("https://api.theporndb.net/jav?external_id=%s", url.QueryEscape(code))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+thePornDBBearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PornBoss/1.0)")

	logging.Info("theporndb request: %s", targetURL)
	resp, err := util.DoRequest(req)
	if err != nil {
		if errors.Is(err, util.ErrCachedNotFound) {
			return "", ResourceNotFonud
		}
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", ResourceNotFonud
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("theporndb: http %d", resp.StatusCode)
	}

	payload, err := decodeThePornDBResponse(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", fmt.Errorf("theporndb: decode response: %w", err)
	}
	if payload == nil || len(payload.Data) == 0 {
		return "", ResourceNotFonud
	}

	for _, item := range payload.Data {
		if !strings.EqualFold(strings.TrimSpace(item.ExternalID), code) {
			continue
		}
		coverURL := strings.TrimSpace(item.Background.Full)
		if coverURL == "" {
			return "", ResourceNotFonud
		}
		return coverURL, nil
	}

	return "", ResourceNotFonud
}

type thePornDBResponse struct {
	Data []thePornDBRecord `json:"data"`
}

type thePornDBRecord struct {
	ExternalID string `json:"external_id"`
	Background struct {
		Full   string `json:"full"`
		Large  string `json:"large"`
		Medium string `json:"medium"`
		Small  string `json:"small"`
	} `json:"background"`
}

func decodeThePornDBResponse(body io.Reader) (*thePornDBResponse, error) {
	var payload thePornDBResponse
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
