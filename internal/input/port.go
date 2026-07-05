package input

import (
	"strconv"
	"strings"
)

// CleanPort trims leading/trailing whitespace from a port string.
func CleanPort(port string) string {
	return strings.TrimSpace(port)
}

// ValidatePort returns true if port is a valid integer in the range 1–65535.
func ValidatePort(port string) bool {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return portNum >= 1 && portNum <= 65535
}
