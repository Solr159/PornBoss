package util

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Accept patterns like "ipx-633", "ipx633", "ipx633_ch", "ipx-714c" (letter suffix ignored).
var CodeRe = regexp.MustCompile(`(?i)([a-z]{2,6})[-_ ]?(\d{2,5})([a-z]{0,2})`)

func ExtractCodeFromName(name string) []string {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	matches := CodeRe.FindAllStringSubmatch(base, -1)
	var out []string
	seen := make(map[string]struct{})
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		suffix := strings.ToUpper(strings.TrimSpace(m[3]))
		number := normalizeNumber(m[2])
		base := fmt.Sprintf("%s-%s", strings.ToUpper(m[1]), number)
		if _, ok := seen[base]; !ok {
			seen[base] = struct{}{}
			out = append(out, base)
		}
		if suffix != "" {
			withSuffix := base + suffix
			if _, ok := seen[withSuffix]; !ok {
				seen[withSuffix] = struct{}{}
				out = append(out, withSuffix)
			}
		}
	}
	return out
}

// normalizeNumber trims leading zeros but keeps at least three digits (padding back if needed).
func normalizeNumber(num string) string {
	num = strings.TrimLeft(num, "0")
	if len(num) == 0 {
		num = "0"
	}
	if len(num) < 3 {
		num = fmt.Sprintf("%0*s", 3, num)
	}
	return strings.ToUpper(num)
}
