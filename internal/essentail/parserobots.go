package essentail

import (
	"net/http"
	"net/url"
	"time"
)

// Function to parse the robots from the input
func ParseRobots(domain string) ([]string, []string, []string, error){
	// Empty lists to store the data parsed
	var disallowed []string
	var sitemaps []string
	var other []string

	// Parses the domain
	baseDomain, err := url.Parse(domain)

	// Checks for errors
	if err != nil{
		return nil, nil, nil, err
	}

	// Resolves the robots path
	robotsURL, err := baseDomain.Parse("/robots.txt")

	// Checks for errors
	if err != nil{
		return nil, nil, nil, err
	}

	// Create the HTTP client with a 10-second timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create the request
	req, err := http.NewRequest("GET", robotsURL.String(), nil)
	if err != nil {
		return nil, nil, nil, err
	}

	// Execute the request
	resp, err := client.Do(req)
	// Catches the error
	if err != nil {
		return nil, nil, nil, err
	}
	// Closes the request when finished
	defer resp.Body.Close()

	// Checks if the reponse is not 200 range and stops the function
	if resp.StatusCode != http.StatusOK {
		// Returns the empty lists
		return disallowed, sitemaps, other, nil
	}
}