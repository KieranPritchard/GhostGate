package main

import (
	"GhostGate/config"
	"GhostGate/internal/commands"
	"GhostGate/internal/input"
	"GhostGate/internal/networking"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// Handle the "init" subcommand before anything else
	if len(os.Args) > 1 && os.Args[1] == "init" {
		if err := config.InitializeConfig(); err != nil {
			log.Fatalf("Failed to initialize: %v", err)
		}
		return
	}

	// Load the configuration file (falls back to built-in defaults on error)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("[!] Warning: Could not load config file, using defaults: %v", err)
	}

	// ---------------------------------------------------------
	// Flag Set Definitions
	// ---------------------------------------------------------

	// ghostgate stage
	stageCmd := flag.NewFlagSet("stage", flag.ExitOnError)
	stagePort := stageCmd.String("p", cfg.DefaultPort, "Port number to host the staging server")
	stageDir := stageCmd.String("d", cfg.DefaultPayloadsDirectory, "Directory path of the staging files")
	stageSource := stageCmd.String("s", "", "Source directory path containing payloads (optional)")
	stageUseTLS := stageCmd.Bool("tls", false, "Enable encrypted HTTPS staging server")
	stageCertFile := stageCmd.String("cert", "", "Path to a custom TLS certificate file")
	stageKeyFile := stageCmd.String("key", "", "Path to a custom TLS private key file")

	// ghostgate upload
	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	uploadPort := uploadCmd.String("p", cfg.DefaultPort, "Port number to host the upload server")
	uploadPath := uploadCmd.String("u", cfg.DefaultURLPath, "URL endpoint path for uploads")
	uploadDest := uploadCmd.String("d", "uploads", "Destination folder to store uploaded files")
	uploadUseTLS := uploadCmd.Bool("tls", false, "Enable encrypted HTTPS upload server")
	uploadCertFile := uploadCmd.String("cert", "", "Path to a custom TLS certificate file")
	uploadKeyFile := uploadCmd.String("key", "", "Path to a custom TLS private key file")

	// ghostgate tunnel
	tunnelCmd := flag.NewFlagSet("tunnel", flag.ExitOnError)
	tunnelPort := tunnelCmd.String("p", cfg.DefaultPort, "Port number to host the local tunnel proxy")
	tunnelTarget := tunnelCmd.String("u", "", "Target URL/endpoint to forward traffic to")
	tunnelUseTLS := tunnelCmd.Bool("tls", false, "Enable encrypted HTTPS tunnel server")
	tunnelCertFile := tunnelCmd.String("cert", "", "Path to a custom TLS certificate file")
	tunnelKeyFile := tunnelCmd.String("key", "", "Path to a custom TLS private key file")

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
		stageCmd.Parse(os.Args[2:])

		cleanPort := input.CleanPort(*stagePort)
		if !input.ValidatePort(cleanPort) {
			log.Fatalf("[!] Invalid port: %s", *stagePort)
		}

		cleanPath := input.CleanFilePath(*stageDir)
		cleanDir, dirValid := input.ValidateFilePath(cleanPath)
		if !dirValid {
			log.Fatalf("[!] Invalid staging directory: %s", *stageDir)
		}

		// The source flag is optional — only validate it when the user provided a value
		cleanSource := ""
		if *stageSource != "" {
			var sourceValid bool
			cleanSource, sourceValid = input.ValidateFilePath(input.CleanFilePath(*stageSource))
			if !sourceValid {
				log.Fatalf("[!] Invalid source directory: %s", *stageSource)
			}
		}

		commands.StagePayloadDirectory(cleanPort, cleanDir, cleanSource, *stageUseTLS, *stageCertFile, *stageKeyFile)

	case "upload":
		uploadCmd.Parse(os.Args[2:])

		cleanPort := input.CleanPort(*uploadPort)
		if !input.ValidatePort(cleanPort) {
			log.Fatalf("[!] Invalid port: %s", *uploadPort)
		}

		if _, err := input.ValidateURL(*uploadPath); err != nil {
			log.Fatalf("[!] Invalid upload path: %v", err)
		}

		commands.StartUploadServer(cleanPort, *uploadPath, *uploadDest, *uploadUseTLS, *uploadCertFile, *uploadKeyFile)

	case "tunnel":
		tunnelCmd.Parse(os.Args[2:])

		if _, err := input.ValidateURL(*tunnelTarget); err != nil {
			log.Fatalf("[!] Invalid tunnel target URL: %v", err)
		}

		cleanPort := input.CleanPort(*tunnelPort)
		if !input.ValidatePort(cleanPort) {
			log.Fatalf("[!] Invalid port: %s", *tunnelPort)
		}

		commands.StartTunnelServer(cleanPort, *tunnelTarget, *tunnelUseTLS, *tunnelCertFile, *tunnelKeyFile)

	case "audit":
		auditCmd.Parse(os.Args[2:])

		if _, err := input.ValidateURL(*auditTarget); err != nil {
			log.Fatalf("[!] Invalid audit target URL: %v", err)
		}

		fmt.Printf("[*] Launching configuration audit against: %s\n", *auditTarget)

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		resp, err := client.Get(*auditTarget)
		if err != nil {
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
