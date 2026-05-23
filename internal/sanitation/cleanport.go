package sanitation

import (
	"log"
	"strconv"
	"strings"
)

// Function to clean file paths
func CleanPort(filePath string) int{
	// Trims the filepath of white space
	trimmedPort := strings.Trim(filePath, "")
	
	// Converts the port to a number
	cleanPort, err := strconv.Atoi(trimmedPort)

	if err != nil {
		// Logs the error
		log.Fatalf("Error Occurred: %s", err)
	}

	// Returns the filepath
	return cleanPort
}