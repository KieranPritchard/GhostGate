package essentail

import (
	"io"
	"net/http"
	"time"
)

// Function for handling tunneling traffic
// Function for handling tunneling traffic - now returns a handler function
func HandleTunnel(target string) http.HandlerFunc {
	return func(writer http.ResponseWriter, reader *http.Request) {
		// Creates a new client
		client := &http.Client{Timeout: 10 * time.Second}
		
		// Stores the request made to the target (using target from the outer function)
		req, err := http.NewRequest(reader.Method, target+reader.RequestURI, reader.Body)
		if err != nil {
			http.Error(writer, "Internal Error", http.StatusInternalServerError)
			return
		}

		// Copies the original headers
		for key, values := range reader.Header {
			for _, value := range values {
				// Copys the header
				req.Header.Add(key, value)
			}
		}

		// Sends a request and gets the error from the request
		resp, err := client.Do(req)
		// Catches the error
		if err != nil {
			// Returns a http error
			http.Error(writer, "Tunnel connection failed", http.StatusBadGateway)
			return
		}
		// Closes when finished
		defer resp.Body.Close()

		// Relays the response back to the orignal sender
		for key, values := range resp.Header {
			for _, value := range values {
				// Writes the header
				writer.Header().Add(key, value)
			}
		}

		// Writes the status code
		writer.WriteHeader(resp.StatusCode)
		// Copies the response body
		io.Copy(writer, resp.Body)
	}
}