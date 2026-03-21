//go:build !darwin

package util

// ClearQuarantineRecursive is a no-op on non-macOS platforms.
func ClearQuarantineRecursive(path string) error {
	return nil
}
