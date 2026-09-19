package input

import (
	"errors"
	"fmt"
	"strings"
)

// type to store the headerss
type Header struct {
	Key string
	Value string
}

// Function to split header and value by space and add to the structure
func PrepareHeaders(headers string) ([]Header, error) {
	// Checks if the headers are passed in
	if len(strings.TrimSpace(headers)) == 0 {
		return nil, errors.New("no headers were entered")
	}

	// Stores the empty arrary
	combinedHeaders := make([]Header, 0)
	
	// Splits each of the strings by their spaces
	separatedHeaders := strings.Fields(headers)

	// Loops over each of the headers
	for _, header := range separatedHeaders {
		// Must contain exactly one '=' relationship: split on the FIRST '=' only
		if !strings.Contains(header, "=") {
			return nil, fmt.Errorf("invalid header %q: missing '=' (expected KEY=VALUE)", header)
		}

		// Gets the header and value
		headerAndValue := strings.SplitN(header, "=", 2)
		key := headerAndValue[0]
		value := headerAndValue[1]

		// Key can't be empty ("=foo" is invalid)
		if key == "" {
			return nil, fmt.Errorf("invalid header %q: empty key", header)
		}

		// Value can't be empty, depending on your needs — decide if you want to allow "Foo="
		if value == "" {
			return nil, fmt.Errorf("invalid header %q: empty value", header)
		}

		// Builds the header and adds it to the combined headers list
		var newHeader Header
		newHeader.Key = key
		newHeader.Value = value

		combinedHeaders = append(combinedHeaders, newHeader)
	}

	return combinedHeaders, nil
}