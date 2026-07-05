package commands

import (
	"GhostGate/internal/networking"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// HandleTunnel returns an http.HandlerFunc that forwards every incoming request
// to target, relays the response headers and body back to the original caller.
func HandleTunnel(target string) http.HandlerFunc {
	return func(writer http.ResponseWriter, reader *http.Request) {
		client := &http.Client{Timeout: 10 * time.Second}

		// Build a new request aimed at the target
		req, err := http.NewRequest(reader.Method, target+reader.RequestURI, reader.Body)
		if err != nil {
			http.Error(writer, "Internal Error", http.StatusInternalServerError)
			return
		}

		// Forward the original request headers
		for key, values := range reader.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Send the proxied request
		resp, err := client.Do(req)
		if err != nil {
			http.Error(writer, "Tunnel connection failed", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Relay the response headers back to the caller
		for key, values := range resp.Header {
			for _, value := range values {
				writer.Header().Add(key, value)
			}
		}

		writer.WriteHeader(resp.StatusCode)
		io.Copy(writer, resp.Body)
	}
}

// StartTunnelServer starts a pivot/tunnel proxy on the given port that forwards
// all requests to target. When useTLS is true the listener is served over HTTPS.
// If certFile and keyFile are provided those are used; otherwise a self-signed
// in-memory certificate is generated automatically.
func StartTunnelServer(port, target string, useTLS bool, certFile, keyFile string) {
	// Set up signal handling so Ctrl+C triggers a clean shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleTunnel(target))

	scheme := "http"
	if useTLS {
		scheme = "https"
	}

	log.Printf("[*] GhostGate Pivot/Tunnel Server active on port %s → %s\n", port, target)
	fmt.Printf("[*] Tunnel listener: curl -X GET %s://%s:%s/<path>\n", scheme, networking.GetOutboundIP(), port)
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
			log.Printf("[-] Tunnel server error: %v\n", err)
			stop <- syscall.SIGTERM
		}
	}()

	<-stop
	fmt.Println("[*] Stopping tunnel server...")
}
