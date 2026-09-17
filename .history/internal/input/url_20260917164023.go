package input

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func PrepareURL(rawURL string) (*url.URL, error)  {
	// Cleans and validates the URL

	// Parses the url
	parsed, err := url.Parse(rawURL)
	if err != nil {
		// Returns nothing
		return nil, err
	}

	// Makes the scheme lowercase
	parsed.Scheme = strings.ToLower(parsed.Scheme)

	// Validates if the scheme is empty
	if parsed.Scheme == "" {
		return nil, errors.New("missing protocol scheme (e.g., http or https)")
	}

	// Checks for unsupported schemes
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported protocol: %s", parsed.Scheme)
	}

	// Checks for a host
	if parsed.Host == "" {
		return nil, errors.New("URL missing host/domain")
	}

	// Returns the parsed url
	return parsed, nil
}