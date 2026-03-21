package util

import (
	"strings"
	"sync"

	"github.com/liuzl/gocc"
)

var (
	goccOnce      sync.Once
	goccConverter *gocc.OpenCC
	goccErr       error
)

// SimplifyChineseName best-effort converts traditional Chinese to simplified.
// If conversion fails, it returns the input.
func SimplifyChineseName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	cc, err := openccConverter()
	if err != nil || cc == nil {
		return value
	}
	simplified, err := cc.Convert(value)
	if err != nil {
		return value
	}
	simplified = strings.TrimSpace(simplified)
	if simplified == "" {
		return value
	}
	return simplified
}

func openccConverter() (*gocc.OpenCC, error) {
	goccOnce.Do(func() {
		goccConverter, goccErr = gocc.New("t2s")
	})
	return goccConverter, goccErr
}
