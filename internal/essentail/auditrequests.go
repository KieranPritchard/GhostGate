package essentail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strings"
	"time"
)

// ANSI Color Escape Codes for Terminal Text
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
)

// AuditRequest connects to a HTTP servere and dumps the structure then terminates it
func AuditRequest(response *http.Response){
	// Checks if there is not a response
	if response == nil {
		// Outputs a nil response cannot be audited
		fmt.Println("[!] Cannot audit a nil response.")

		// Returns nothing
		return
	}
	// Closes the response when done
	defer response.Body.Close()
	
	// 1. Format and Colorize Headers
	fmt.Printf("\n%s=== [TARGET HTTP SERVER AUDIT: %s] ===%s\n", ColorCyan, time.Now().Format("15:04:05"), ColorReset)
	fmt.Printf("%sProto:%s    %s\n", ColorBlue, ColorReset, response.Proto)
	fmt.Printf("%sStatus:%s   %s\n\n", ColorBlue, ColorReset, formatStatus(response.Status))

	// Outputs the header section
	fmt.Printf("%s--- HEADERS ---%s\n", ColorYellow, ColorReset)
	
	// Loops over the keys and the values
	for key, values := range response.Header {
		// Outputs key and value
		fmt.Printf("  %s%s:%s %s\n", ColorCyan, key, ColorReset, strings.Join(values, ", "))
	}

	// Reads and formats the body
	bodyBytes, err := io.ReadAll(response.Body)

	// Checks for errors
	if err != nil {
		fmt.Printf("\n[!] Failed to read response body: %v\n", err)
		return
	}

	// Checks if the lengtg is greater than zero
	if len(bodyBytes) > 0 {
			// Outputs a header
			fmt.Printf("\n%s--- BODY ---%s\n", ColorYellow, ColorReset)
			
			// Gets the content type
			contentType := response.Header.Get("Content-Type")
			
			// Checks for a json application
			if strings.Contains(contentType, "application/json") {
				// Formats the json
				var prettyJSON bytes.Buffer
				
				// Formats the json with proper indents
				if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
					fmt.Println(prettyJSON.String())
				} else {
					// Fallback to raw if JSON parsing fails
					fmt.Println(string(bodyBytes))
				}
			} else {
				// It's HTML, Plain Text, etc.
				fmt.Println(string(bodyBytes))
			}
		}

		fmt.Printf("%s=====================================================%s\n", ColorCyan, ColorReset)
}

// Helper to make 2xx status green and 4xx/5xx status red
func formatStatus(status string) string {
	// Checks if the string is in the 200 range
	if strings.HasPrefix(status, "2") {
		return ColorGreen + status + ColorReset
	}

	// Checks if the string is in the 400 or 500 range 
	if strings.HasPrefix(status, "4") || strings.HasPrefix(status, "5") {
		return ColorRed + status + ColorReset
	}
	// Returns the colour
	return ColorYellow + status + ColorReset
}

// PrettyPrintHTML takes a raw HTML string and returns a cleanly indented version
func prettyPrintHTML(htmlStr string) string {
	// Creates a new buffer
	var buf bytes.Buffer

	// Reads in the dom
	dom := html.NewTokenizer(strings.NewReader(htmlStr))
	
	// Tracks the indent
	indent := 0

	// Helper to write spaces for indentation
	writeIndent := func(level int) {
		// Loops over each of the levels
		for i := 0; i < level; i++ {
			buf.WriteString("  ") // 2 spaces per indent level
		}
	}

	// Infinately loops
	for {
		// Creates a new dom token type
		tokenType := dom.Next()

		// Checks if the token type is an error
		if tokenType == html.ErrorToken {
			// Checks if the dom is an error and end of file and breaks
			if dom.Err() == io.EOF {
				break
			}
			return htmlStr // Return raw HTML if parsing fails entirely
		}

		// Creates another token
		token := dom.Token()

		// Checks the token type
		switch tokenType {
			// Checks if there is a start tag 
			case html.StartTagToken:
				// Don't indent if it's inline text structural tagging
				writeIndent(indent)

				// Strings the token and a new line to the buffer
				buf.WriteString(token.String() + "\n")
				// Self-closing tags (like <img/> or <input>) shouldn't increase indentation
				if !isSelfClosing(token.Data) {
					indent++
				}
			
			// Checks for end tag tokens
			case html.EndTagToken:
				// Lowers the input
				indent--

				// Writes the indent
				writeIndent(indent)
				buf.WriteString(token.String() + "\n")
			
			// Checks for self closing tags
			case html.SelfClosingTagToken:
				writeIndent(indent)
				buf.WriteString(token.String() + "\n")
			
			// Checks for the text
			case html.TextToken:
				text := strings.TrimSpace(token.Data)
				if len(text) > 0 {
					writeIndent(indent)
					buf.WriteString(text + "\n")
				}
		}
	}

	// Returns a buffer
	return buf.String()
}

// Quick check for common HTML tags that don't close traditionally
func isSelfClosing(tag string) bool {
	// Checks for the tag
	switch tag {
	case "meta", "link", "br", "hr", "img", "input":
		// returns true
		return true
	}
	// returns false
	return false
}