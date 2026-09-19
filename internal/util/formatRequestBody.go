package util

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"mime"
	"net/url"
	"strings"
)

// formatBody formats response body content based on its media type
func FormatBody(contentType string, bodyBytes []byte) string {
	// Parses the base media type, ignoring parameters like charset
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil && contentType != "" {
		mediaType = contentType
	}

	switch {
	// JSON
	case mediaType == "application/json" || strings.HasSuffix(mediaType, "+json"):
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
			return prettyJSON.String()
		}

	// XML — validated for well-formedness before falling back to raw output
	case mediaType == "application/xml" || mediaType == "text/xml" || strings.HasSuffix(mediaType, "+xml"):
		if err := xml.Unmarshal(bodyBytes, new(interface{})); err == nil {
			return string(bodyBytes)
		}

	// 3. Form URL-Encoded
	case mediaType == "application/x-www-form-urlencoded":
		if values, err := url.ParseQuery(string(bodyBytes)); err == nil {
			var b strings.Builder
			for key, val := range values {
				b.WriteString(fmt.Sprintf("%s = %v\n", key, val))
			}
			return strings.TrimSpace(b.String())
		}

	// 4. HTML — extracts security-relevant hits instead of dumping raw markup
	case mediaType == "text/html":
		return ExtractHTMLSecurityHits(bodyBytes)
	}

	// Falls back to raw string output if parsing/formatting fails
	return string(bodyBytes)
}
