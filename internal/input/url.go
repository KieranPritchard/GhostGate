package input

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateURL parses rawURL and verifies it has a valid http/https scheme and a host.
// Returns the parsed URL or a descriptive error.
func ValidateURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse URL: %w", err)
	}

	if parsed.Scheme == "" {
		return nil, fmt.Errorf("missing protocol scheme (e.g., http or https)")
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported protocol: %s", parsed.Scheme)
	}

	if parsed.Host == "" {
		return nil, fmt.Errorf("URL missing host/domain")
	}

	return parsed, nil
}
