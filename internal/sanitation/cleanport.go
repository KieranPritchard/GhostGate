package sanitation

import (
	"strings"
)

// CleanPort trims leading/trailing whitespace from a port string.
func CleanPort(port string) string {
	return strings.TrimSpace(port)
}