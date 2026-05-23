package sanitation

import (
	"strings"
)

// Function to clean port numbers
func CleanPort(port string) string{
	// Trims the filepath of white space
	trimmedPort := strings.Trim(port, "")

	// Returns the filepath
	return trimmedPort
}