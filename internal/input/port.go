package input

import (
	"errors"
	"strconv"
	"strings"
)

func PreparePort(port string) (string, error)  {
	// Cleans and validates the ports
	
	// Removes trailing spaces
	port = strings.TrimSpace(port)

	// Converts the port to a number
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return "", err
	}

	// Checks if in correct range
	if portNum < 1 || portNum > 65535 {
		return "", errors.New("port number must be between 1 & 65535")
	}

	// Returns the port
	return port, nil
}