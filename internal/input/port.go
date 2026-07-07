package input

import (
	"errors"
	"strconv"
	"strings"
)

// CleanPort trims leading/trailing whitespace from a port string.
func CleanPort(port string) string {
	return strings.TrimSpace(port)
}

// ValidatePort returns true if port is a valid integer in the range 1–65535.
func ValidatePort(port string) (error) {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	if portNum < 1 || portNum > 65535 {
		return errors.New("Port number must be between 1 & 65535")
	}

	return nil
}
