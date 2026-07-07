package main

import (
	"GhostGate/config"
	"GhostGate/internal/commands"
	"GhostGate/internal/input"
	"GhostGate/internal/logger"
	"GhostGate/internal/networking"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// Creates a new context
	ctx := context.Background()

	// Creates a log file
	logFile, err := os.OpenFile("ghostgate.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	// Creates a new logger
	logger.New(logger.Config{
		Level: "INFO",
		Format: logger.FormatJSON,
		AddSource: true,
		Output: logFile,
	})

	// Logs the app has started
	logger.Info(ctx, "service starting")

	// Handle the "init" subcommand before anything else
	if len(os.Args) > 1 && os.Args[1] == "init" {
		// Logs that the initaliseion has started
		logger.Info(ctx, "initalisation starting")
		
		if err := config.InitializeConfig(); err != nil {
			// Outputs the initalsation has failed
			fmt.Printf("Failed to initialize: %v\n", err)

			// Logs the init command has failed
			logger.Error(ctx, "Initialization failed", err)
		}
		
		// Logs the inistalsation has finished
		logger.Info(ctx, "initalisation finished")
		return
	}

	// Load the configuration file (falls back to built-in defaults on error)
	cfg, err := config.LoadConfig()

	// Logs the config is being loaded
	logger.Info(ctx, "Config is being loaded")

	if err != nil {
		// Outputs the config file could not be loaded
		fmt.Printf("[!] Warning: Could not load config file, using defaults: %v", err)

		// Logs the config file could not be loaded
		logger.Error(ctx, "Could not load config file, using defaults")
	}

	// ---------------------------------------------------------
	// Flag Set Definitions
	// ---------------------------------------------------------

	// ghostgate stage
	stageCmd := flag.NewFlagSet("stage", flag.ExitOnError)
	stagePort := stageCmd.String("p", cfg.DefaultPort, "Port number to host the staging server")
	stageDir := stageCmd.String("d", cfg.DefaultPayloadsDirectory, "Directory path of the staging files")
	stageSource := stageCmd.String("s", "", "Source directory path containing payloads (optional)")
	stageUseTLS := stageCmd.Bool("tls", cfg.DefaultTLSEnabled, "Enable encrypted HTTPS staging server")
	stageCertFile := stageCmd.String("cert", cfg.DefaultTLSCertFile, "Path to a custom TLS certificate file")
	stageKeyFile := stageCmd.String("key", cfg.DefaultTLSKeyFile, "Path to a custom TLS private key file")

	// ghostgate upload
	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	uploadPort := uploadCmd.String("p", cfg.DefaultPort, "Port number to host the upload server")
	uploadPath := uploadCmd.String("u", cfg.DefaultURLPath, "URL endpoint path for uploads")
	uploadDest := uploadCmd.String("d", cfg.DefaultUploadsDirectory, "Destination folder to store uploaded files")
	uploadUseTLS := uploadCmd.Bool("tls", cfg.DefaultTLSEnabled, "Enable encrypted HTTPS upload server")
	uploadCertFile := uploadCmd.String("cert", cfg.DefaultTLSCertFile, "Path to a custom TLS certificate file")
	uploadKeyFile := uploadCmd.String("key", cfg.DefaultTLSKeyFile, "Path to a custom TLS private key file")

	// ghostgate tunnel
	tunnelCmd := flag.NewFlagSet("tunnel", flag.ExitOnError)
	tunnelPort := tunnelCmd.String("p", cfg.DefaultPort, "Port number to host the local tunnel proxy")
	tunnelTarget := tunnelCmd.String("u", "", "Target URL/endpoint to forward traffic to")
	tunnelUseTLS := tunnelCmd.Bool("tls", cfg.DefaultTLSEnabled, "Enable encrypted HTTPS tunnel server")
	tunnelCertFile := tunnelCmd.String("cert", cfg.DefaultTLSCertFile, "Path to a custom TLS certificate file")
	tunnelKeyFile := tunnelCmd.String("key", cfg.DefaultTLSKeyFile, "Path to a custom TLS private key file")

	// ghostgate audit
	auditCmd := flag.NewFlagSet("audit", flag.ExitOnError)
	auditTarget := auditCmd.String("u", "", "Target URL/endpoint to audit")

	// Require a subcommand
	if len(os.Args) < 2 {
		fmt.Println("[!] Usage: ghostgate <stage|upload|tunnel|audit|init> [flags]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "stage":
		// Logs the commands are being parsed
		logger.Info(ctx, "Parsing commands for 'stage'")

		// Parses the stage command arguements
		stageCmd.Parse(os.Args[2:])

		// Logs the ports are being cleaned
		logger.Info(ctx, "Cleaning entered port", *stagePort)

		// Cleans the port number
		cleanPort := input.CleanPort(*stagePort)

		// Logs the validation has started
		logger.Info(ctx, "Starting validation on cleaned port", cleanPort)

		// Validates the port number
		if !input.ValidatePort(cleanPort) {
			// Logs the port is invalid
			logger.Error(ctx, "Validation failed on port", cleanPort)

			// Outputs the port is invalid
			fmt.Printf("[!] Invalid port: %s", *stagePort)
		}

		// Logs the file path is being cleaned
		logger.Info(ctx, "Cleaning path for stage directory", *stageDir)
		
		// Cleans the path for the staging directory
		cleanPath := input.CleanFilePath(*stageDir)

		// Logs the directory is being cleaned
		logger.Info(ctx, "Validating the clean staging directory", cleanPath)

		// Checks if the clean directory is valid
		cleanDir, dirValid := input.ValidateFilePath(cleanPath)
		if !dirValid {
			// Logs the validation has failed
			logger.Info(ctx, "Validation of the staging directory has failed", cleanPath)
			
			// Outputs the staging directory is invalid
			fmt.Printf("[!] Invalid staging directory: %s", *stageDir)
		}

		// The source flag is optional — only validate it when the user provided a value
		cleanSource := ""

		// Checks if a stage source was entered
		if *stageSource != "" {
			// Stores if the source is valid
			var sourceValid bool
			
			// Logs if the source directory is being validated
			logger.Info(ctx, "Validating source directory for the staging", *stageSource)

			// Validates the source path
			cleanSource, sourceValid = input.ValidateFilePath(input.CleanFilePath(*stageSource))
			if !sourceValid {
				// logs the source path is invalid
				logger.Error(ctx, "Invalid source directory", *stageSource)

				// Outputs the source is invalid
				fmt.Printf("[!] Invalid source directory: %s", *stageSource)
			}
		}

		// Runs the stage payload directory function
		commands.StagePayloadDirectory(cleanPort, cleanDir, cleanSource, *stageUseTLS, *stageCertFile, *stageKeyFile)

	case "upload":
		// Logs the commands are being parsed
		logger.Info(ctx, "Parsing commands for 'upload'")

		// Parses the arguements for the upload command
		uploadCmd.Parse(os.Args[2:])
		
		// Logs the ports are being cleaned
		logger.Info(ctx, "Cleaning entered port", *uploadPort)
		
		// Cleans the port
		cleanPort := input.CleanPort(*uploadPort)

		// Logs the validation has started
		logger.Info(ctx, "Starting validation on cleaned port", cleanPort)

		// Validating the clean port
		if !input.ValidatePort(cleanPort) {
			// Logs the port is invalid
			logger.Error(ctx, "Validation failed on port", cleanPort)
			
			// Prints the port is invalid
			fmt.Printf("[!] Invalid port: %s", *uploadPort)
		}

		// Logs validation has started
		logger.Info(ctx, "Validation has started on path", *uploadPath)

		// CHecks if there is an error
		if _, err := input.ValidateURL(*uploadPath); err != nil {
			// Logs the url is invalid
			logger.Error(ctx, "Upload path is invalid", *uploadPath)

			// Prints the path is invalid
			fmt.Printf("[!] Invalid upload path: %v", err)
		}

		commands.StartUploadServer(cleanPort, *uploadPath, *uploadDest, *uploadUseTLS, *uploadCertFile, *uploadKeyFile)

	case "tunnel":
		// Logs the commands are being parsed
		logger.Info(ctx, "Parsing commands for 'tunnel'")

		// Parses the tunnel command
		tunnelCmd.Parse(os.Args[2:])

		// Logs validation has start
		logger.Info(ctx, "Validation started on the url", *tunnelTarget)

		// Check if there is an error
		if _, err := input.ValidateURL(*tunnelTarget); err != nil {
			// Logs the validation has failed
			logger.Error(ctx, "Validation failed on tunnel target", *tunnelTarget)
			
			// Outputs the target is invalid
			fmt.Printf("[!] Invalid tunnel target URL: %v", err)
		}

		// Logs the port is being cleaned
		logger.Info(ctx, "Cleaning has started on the port", *tunnelPort)

		// Cleans the port entered
		cleanPort := input.CleanPort(*tunnelPort)
		if !input.ValidatePort(cleanPort) {
			// Logs the port is invalid
			logger.Error(ctx, "Port is invalid", *tunnelPort)

			// Ouputs the port is invalid
			fmt.Printf("[!] Invalid port: %s", *tunnelPort)
		}

		commands.StartTunnelServer(cleanPort, *tunnelTarget, *tunnelUseTLS, *tunnelCertFile, *tunnelKeyFile)

	case "audit":
		// Logs the commands are being parsed
		logger.Info(ctx, "Parsing commands for 'audit'")

		// Parses the audit commands
		auditCmd.Parse(os.Args[2:])

		// Logs the url validation has started
		logger.Info(ctx, "Starting validation on the url")
		
		// Checks if the url is valid
		if _, err := input.ValidateURL(*auditTarget); err != nil {
			// Logs the url is incorrect
			logger.Error(ctx, "Invalid url", *auditTarget)

			// Outputs the url is incorrect
			fmt.Printf("[!] Invalid audit target URL: %v", err)
		}

		// Prints the configuration is starting
		fmt.Printf("[*] Launching configuration audit against: %s\n", *auditTarget)

		// Creates a http client
		client := &http.Client{
			// Sets timeout to 10 seconds
			Timeout: 10 * time.Second,
		}

		// Gets the response
		resp, err := client.Get(*auditTarget)
		if err != nil {
			// Logs the response failed
			logger.Error(ctx, "Connection failed", err)
			log.Fatalf("[!] Connection failed: %v\n", err)
		}
		defer resp.Body.Close()

		commands.AuditRequest(resp)

	default:
		fmt.Printf("[!] Unknown command: %s\n", os.Args[1])
		fmt.Println("[!] Usage: ghostgate <stage|upload|tunnel|audit|init> [flags]")

		// Print outbound IP for quick reference
		fmt.Printf("[*] Local IP: %s\n", networking.GetOutboundIP())
		os.Exit(1)
	}
}