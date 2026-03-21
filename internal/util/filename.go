package util

import "net/url"

// FileNameForFingerprint returns a filesystem-safe thumbnail name for a fingerprint.
func FileNameForFingerprint(fingerprint string) string {
	if fingerprint == "" {
		return ""
	}
	return url.PathEscape(fingerprint) + ".jpg"
}
