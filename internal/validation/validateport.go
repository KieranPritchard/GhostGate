package validation

import (
	"log"
	"strconv"
)

// Function to validate port numbers
func ValidatePort(port string) bool{
	// Performs a length check on the stage directory port
	if len(port) > 5 {
		// Logs the error
		log.Fatalf("Invalid port: %s. Port must be a length of 5 or lower", port)

		// Returns false
		return false
	}

	// Checks if the port flag is a number
	_, err := strconv.Atoi(port)
	
	// CHecks for an error
	if err != nil {
		// Logs the error
		log.Fatalf("Invalid port: %s. Port must be a number", port)
		
		// Returns false
		return false
	}

	// Converts the port number to perform a range check
	port_num, err := strconv.Atoi(port)

	// Performs the range check
	if err != nil || port_num < 1 || port_num > 65535 {
		// Logs the error
		log.Fatalf("Invalid port: %s. Port must be a number between 1 and 65535", port)
	
		// Returns false
		return false
	}

	// Returns true
	return true
}