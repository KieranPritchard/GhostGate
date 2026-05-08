package essentail

import (
	"GhostGate/internal/filesystem"
	"GhostGate/internal/networking"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Function for staging the payload directory
func StagePayloadDirectory(port string, stagingDir string, sourceDir string)  {
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
		files, err := ioutil.ReadDir(sourceDir)

		// Catches the errors
		if err != nil {
			// Logs the error
			log.Fatal(err)
		}

		// Loops over the files
		for _, file := range files {
			// Gets the names of the files
			name := file.Name()

			// Creates the paths
			srcPath := filepath.Join(sourceDir, name)
			dstPath := filepath.Join(stagingDir, name)
			
			// Copies the files
			filesystem.CopyFile(srcPath, dstPath)
		}
	}

	// Creates the file server handler
	fileServer := http.FileServer(http.Dir(stagingDir))

	// Outputs information
	fmt.Printf("[*] Go Payload Staging Server running on port %s\n", port)
	fmt.Printf("[*] Serving files from: %s\n", stagingDir)
	fmt.Printf("[*] Target download example: curl http://%s:%s/%s/file\n", networking.GetOutboundIP(), port, stagingDir)

	// Starts the server
	err := http.ListenAndServe(":"+port, fileServer)

	// Checks if there is an error
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}