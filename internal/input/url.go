package input

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// CleanURL parses a raw url and preps it for the validateURL function
func CleanURL(rawURL string) *url.URL {
	// Parses the url
	parsed, err := url.Parse(rawURL)
	if err != nil {
		fmt.Printf("Could not parse URL: %e", err)

		// Returns nothing
		return nil
	}

	// Makes the scheme lowercase
	parsed.Scheme = strings.ToLower(parsed.Scheme)

	// Returns the parsed url
	return parsed
}

// ValidateURL parses rawURL and verifies it has a valid http/https scheme and a host.
// Returns the parsed URL or a descriptive error.
func ValidateURL(URL *url.URL) (error) {
	// Validates if the scheme is empty
	if URL.Scheme == "" {
		return fmt.Errorf("missing protocol scheme (e.g., http or https)")
	}

	// Checks for unsupported schemes
	if URL.Scheme != "http" && URL.Scheme != "https" {
		return fmt.Errorf("unsupported protocol: %s", URL.Scheme)
	}

	// Checks for a host
	if URL.Host == "" {
		return errors.New("URL missing host/domain")
	}

	return nil
}
