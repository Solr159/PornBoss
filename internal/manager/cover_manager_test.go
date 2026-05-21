package manager

import (
	"net/http"
	"strings"
	"testing"
)

func TestSetCoverDownloadHeadersForJavBus(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://www.javbus.com/pics/cover/c85j_b.jpg", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	setCoverDownloadHeaders(req)

	if got := req.Header.Get("Referer"); got != "https://www.javbus.com/" {
		t.Fatalf("Referer = %q, want javbus referer", got)
	}
	if got := req.Header.Get("Cookie"); !strings.Contains(got, "age=verified") {
		t.Fatalf("Cookie = %q, want age verified cookie", got)
	}
	if got := req.Header.Get("User-Agent"); !strings.Contains(got, "Chrome/") {
		t.Fatalf("User-Agent = %q, want browser user agent", got)
	}
}

func TestSetCoverDownloadHeadersForDMM(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://pics.dmm.co.jp/digital/video/avsa00388/avsa00388pl.jpg", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	setCoverDownloadHeaders(req)

	if got := req.Header.Get("Referer"); got != "https://www.dmm.co.jp/" {
		t.Fatalf("Referer = %q, want dmm referer", got)
	}
	if got := req.Header.Get("User-Agent"); !strings.Contains(got, "Chrome/") {
		t.Fatalf("User-Agent = %q, want browser user agent", got)
	}
}
