package sanitation

import (
	"path/filepath"
	"strings"
)

// CleanFilePath trims leading/trailing whitespace and resolves the filepath.
func CleanFilePath(filePath string) string {
	// Trim whitespace from the filepath
	trimmedPath := strings.TrimSpace(filePath)

	// Lexically clean the filepath
	cleanPath := filepath.Clean(trimmedPath)

	return cleanPath
}