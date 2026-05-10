package sanitation

import "path/filepath"

// Function to clean file paths
func CleanFilePath(filePath string) string{
	// Cleans the filepath
	cleanPath := filepath.Clean(filePath)

	// Returns the filepath
	return cleanPath
}