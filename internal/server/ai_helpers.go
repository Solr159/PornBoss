package server

import (
	"context"
	"fmt"
	"os"
	"strings"

	dbpkg "pornboss/internal/db"
)

func deepseekCredentials(ctx context.Context) (apiKey, baseURL string, err error) {
	if k := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")); k != "" {
		apiKey = k
	}
	if u := strings.TrimSpace(os.Getenv("DEEPSEEK_BASE_URL")); u != "" {
		baseURL = u
	}
	cfg, err := dbpkg.ListConfig(ctx)
	if err != nil {
		return "", "", err
	}
	if apiKey == "" {
		apiKey = strings.TrimSpace(cfg["deepseek_api_key"])
	}
	if baseURL == "" {
		baseURL = strings.TrimSpace(cfg["deepseek_base_url"])
	}
	return apiKey, baseURL, nil
}

func sanitizeConfigForClient(cfg map[string]string) map[string]string {
	if cfg == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(cfg)+1)
	for k, v := range cfg {
		if k == "deepseek_api_key" {
			continue
		}
		out[k] = v
	}
	if strings.TrimSpace(cfg["deepseek_api_key"]) != "" {
		out["deepseek_api_key_set"] = "1"
	} else {
		out["deepseek_api_key_set"] = "0"
	}
	return out
}

func extractJSONObjectFromLLM(content string) (string, error) {
	s := strings.TrimSpace(content)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimPrefix(s, "json")
		s = strings.TrimSpace(s)
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = strings.TrimSpace(s[:idx])
		}
	}
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end <= start {
		return "", fmt.Errorf("no json object in model output")
	}
	return s[start : end+1], nil
}
