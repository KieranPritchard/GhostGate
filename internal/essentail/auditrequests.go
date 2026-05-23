package essentail

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"
)

// AuditRequest connects to a HTTP servere and dumps the structure then terminates it
func AuditRequest(response *http.Response){
	// DumpResponse extracts the raw protocol bytes retuned by the target server.
	// Setting the second parameter to 'true' includes the HTML/data body.
	responseDump, err := httputil.DumpResponse(response, true)

	// Checks for any errors
	if err != nil {
		// Prints an error
		fmt.Printf("[!] Failed to dump server response: %v\n", err)
		return
	}

	// Print the raw headers and server flags directly to your terminal
	fmt.Printf("\n=== [TARGET HTTP SERVER AUDIT: %s] ===\n", time.Now().Format("15:04:05"))
	fmt.Println(string(responseDump))
	fmt.Println("=====================================================")
}