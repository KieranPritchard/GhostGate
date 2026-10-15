package input

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
)

// Cleans and validates the file paths
func PrepareFilePath(path string) (string, error) {
	// Checks if the path is empty
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", errors.New("file path must not be empty")
	}

	// Cleans the file path and trims the space
	path = filepath.Clean(strings.TrimSpace(path))

	// Checks for at least one alphabetical character
	match, err := regexp.MatchString(`[[:alpha:]]`, path)
	if err != nil {
		return "", errors.New("file path does not match the expected format")
	}

	if !match {
		return "", errors.New("path is not the correct format")
	}

	return path, nil
}
