package util

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/mattn/go-ieproxy"
	"pornboss/internal/common/logging"
)

var (
	proxyOnce sync.Once
	proxyFunc func(*http.Request) (*url.URL, error)
	proxyPort atomic.Value // stores *url.URL
)

// DetectProxyFunc returns a proxy function that prefers manual override, then env/system
// proxies provided by go-ieproxy (env has priority inside). The result is cached.
func DetectProxyFunc() func(*http.Request) (*url.URL, error) {
	proxyOnce.Do(func() {
		proxyFunc = resolveProxy()
	})
	return proxyFunc
}

// SetProxyPort configures the local proxy port. Use <=0 to disable.
func SetProxyPort(port int) {
	if port <= 0 {
		proxyPort.Store((*url.URL)(nil))
		logging.Info("proxy: cleared configured port")
		return
	}
	u := &url.URL{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", port)}
	proxyPort.Store(u)
	logging.Info("proxy: using configured port %s", u.Redacted())
}

// SetProxyPortFromString parses a port string and configures the local proxy port.
func SetProxyPortFromString(raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		SetProxyPort(0)
		return
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port <= 0 || port > 65535 {
		SetProxyPort(0)
		return
	}
	SetProxyPort(port)
}

func resolveProxy() func(*http.Request) (*url.URL, error) {
	systemProxy := ieproxy.GetProxyFunc()
	return func(req *http.Request) (*url.URL, error) {
		if u := loadProxyOverride(); u != nil {
			return u, nil
		}
		if systemProxy != nil {
			return systemProxy(req)
		}
		return nil, nil
	}
}

func loadProxyOverride() *url.URL {
	val := proxyPort.Load()
	if val == nil {
		return nil
	}
	if u, ok := val.(*url.URL); ok {
		return u
	}
	return nil
}
