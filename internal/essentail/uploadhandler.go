package essentail

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Function to handle file uploads, accepting a dynamic folder name as an argument
func UploadHandler(exfilDir string) http.HandlerFunc {
	return func(writer http.ResponseWriter, reader *http.Request) {
		// Checks if the correct method was being used
		if reader.Method != http.MethodPost {
			// Returns a error message
			http.Error(writer, "Use POST to exfiltrate data", http.StatusMethodNotAllowed)
			return
		}

		// Tries to create the the directory to store the files
		if err := os.MkdirAll(exfilDir, 0755); err != nil {
			log.Printf("[-] Failed to create storage directory: %v", err)
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Retrieves a filename from a custom header if not uses a default
		filename := reader.Header.Get("X-File-Name")

		// Checks if the filename is a empty string
		if filename == "" {
			// Creates the new file name
			filename = "exfil_data.bin"
		} else {
			// Creates the clean file name
			filename = filepath.Base(filename)
		}

		// Creates the destination path
		dstPath := filepath.Join(exfilDir, filename)

		// Create the destination file
		dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

		// Checks for and logs errors
		if err != nil {
			log.Printf("[-] Failed to create file %s: %v", dstPath, err)
			http.Error(writer, "Failed to create destination file", http.StatusInternalServerError)
			return
		}
		// the file descriptor is closes when the HTTP handler function returns
		defer dst.Close()

		// Streams the request body directly to disk
		bytesCopied, err := io.Copy(dst, reader.Body)

		// Checks for and logs the errors
		if err != nil {
			log.Printf("[!] Error during data transfer from %s: %v", reader.RemoteAddr, err)
			http.Error(writer, "Error saving file data", http.StatusInternalServerError)
			return
		}

		// Logs success status
		log.Printf("[+] Data Transfer Successful: %d bytes received from %s saved as %s", bytesCopied, reader.RemoteAddr, filename)
		// Sends back an created status
		writer.WriteHeader(http.StatusCreated)
	}
}