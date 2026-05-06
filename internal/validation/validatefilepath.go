package validation

import (
	"log"
	"path/filepath"
	"regexp"
)

// Function to validate file paths
func ValidateFilePath(filePath string) (string, bool){

	// Checks if the path is empty
	if filePath == "" {
		// Logs the error
		log.Fatalf("Invaild format: %s. Directory must not be empty", filePath)

		// Returns none and false
		return "", false
	}

	// Cleans the filepath
	cleanPath := filepath.Clean(filePath)
	
	// Checks if the file path contains letters
	match, _ := regexp.MatchString(`[[:alpha:]]`, filePath)

	// Checks if there is not a match
	if !match {

		// Logs the error
		log.Fatalf("Invaild format: %s. Directory must include some letters", filePath)
	
		// Returns none and false
		return "", false
	}

	// Returns the data
	return cleanPath, true
}