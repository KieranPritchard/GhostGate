package validation

import (
	"strconv"
)

// ValidatePort returns true if the port string is a valid integer in the range 1–65535.
func ValidatePort(port string) bool {
	// Parse the port into an integer
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return false
	}

	// Check it is within the valid TCP/UDP port range
	if portNum < 1 || portNum > 65535 {
		return false
	}

	return true
}