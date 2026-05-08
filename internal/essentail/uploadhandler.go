package essentail

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Function to handle file uploads
func UploadHandler(writer http.ResponseWriter, reader *http.Request) {
	// Checks if the correct method was being used
	if reader.Method != http.MethodPost {
		// Returns a error message
		http.Error(writer, "Use POST to exfiltrate data", http.StatusMethodNotAllowed)
		return
	}

	// Creates a directory to store the data send to the server
	exfilDir := "./exfiltrated_data"
	os.MkdirAll(exfilDir, os.ModePerm)

	// Retrieves a filename from a custom header if not uses a default
	filename := reader.Header.Get("X-File-Name")

	// Checks if the filename is a empty string
	if filename == "" {
		// Creates the new file name
		filename = "exfil_data.bin"
	}

	// Creates the destination path
	dstPath := filepath.Join(exfilDir, filename)

	// Creates the file and gets the error
	dst, err := os.Create(dstPath)

	// Checks for errors
	if err != nil {
		// Returns an http header
		http.Error(writer, "Failed to create destination file", http.StatusInternalServerError)
		return
	}

	// Closes the path when done
	defer dst.Close()

	// Stream the body directly to disk to handle large "exfiltrations" efficiently
	bytesCopied, err := io.Copy(dst, reader.Body)

	// Checks for an error
	if err != nil {
		// Logs the error
		log.Printf("[!] Error during exfiltration from %s: %v", reader.RemoteAddr, err)
		return
	}

	// Outputs a success message 
	log.Printf("[+] Exfiltration Successful: %d bytes received from %s saved as %s", bytesCopied, reader.RemoteAddr, filename)
	// Writes the header
	writer.WriteHeader(http.StatusCreated)
}
