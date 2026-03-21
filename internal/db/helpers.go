package db

import "strings"

func normalizeTagNames(names []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(names))
	for _, name := range names {
		cleaned := strings.TrimSpace(name)
		if cleaned == "" {
			continue
		}
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		result = append(result, cleaned)
	}
	return result
}

func uniqueInt64s(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func normalizeNames(names []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(names))
	for _, name := range names {
		cleaned := strings.TrimSpace(name)
		if cleaned == "" {
			continue
		}
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		result = append(result, cleaned)
	}
	return result
}
