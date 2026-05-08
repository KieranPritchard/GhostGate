package validation

import (
	"fmt"
	"net/url"
	"strings"
)

// Function to validate url
func ValidateURL(rawURL string) (*url.URL, error){
	// Parses the url
	url, err := url.Parse(rawURL)

	// Checks for an error
	if err != nil{
		// Returns that the url could not be parsed
		return nil, fmt.Errorf("Could not parse URL: %w", err)
	}

	// Checks for url scheme
	if url.Scheme == "" {
		// Returns missing schema
		return nil, fmt.Errorf("missing protocol scheme (e.g., http or https)")
	}

	// Makes the schem lowercase
	url.Scheme = strings.ToLower(url.Scheme)

	// Checks if the scheme is not http or https
	if url.Scheme != "http" && url.Scheme != "https" {
		// Returns a protocol error
		return nil, fmt.Errorf("unsupported protocol: %s", url.Scheme)
	}

	// Checks if there is a host name
	if url.Host == "" {
		// Returns nil and a format error
		return nil, fmt.Errorf("URL missing host/domain")
	}

	// Returns an url and an error
	return url, nil
}