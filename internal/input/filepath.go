package input

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
)

func PrepareFilePath(path string) (string, error)  {
	// Cleans and validates the file paths

	// Cleans the file path and trims the space
	path = filepath.Clean(strings.TrimSpace(path))

	// Checks if the path is empty
	if path == "" {
		return "", errors.New("file path must not be empty")
	}

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