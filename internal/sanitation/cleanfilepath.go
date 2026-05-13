package sanitation

import (
	"path/filepath"
	"strings"
)

// Function to clean file paths
func CleanFilePath(filePath string) string{
	// Trims the filepath of white space
	trimmedPath := strings.Trim(filePath, "")
	
	// Cleans the filepath
	cleanPath := filepath.Clean(trimmedPath)

	// Returns the filepath
	return cleanPath
}