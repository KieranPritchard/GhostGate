package input

import (
	"path/filepath"
	"regexp"
	"strings"
)

// CleanFilePath trims leading/trailing whitespace and lexically cleans the path.
func CleanFilePath(path string) string {
	return filepath.Clean(strings.TrimSpace(path))
}

// ValidateFilePath checks that filePath is non-empty and contains at least one letter.
// Returns the path and true on success, or an empty string and false on failure.
func ValidateFilePath(filePath string) (string, bool) {
	if filePath == "" {
		return "", false
	}

	// Require at least one alphabetical character
	match, _ := regexp.MatchString(`[[:alpha:]]`, filePath)
	if !match {
		return "", false
	}

	return filePath, true
}
