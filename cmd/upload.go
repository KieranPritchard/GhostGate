package cmd

import (
	"GhostGate/internal/commands"
	"GhostGate/internal/input"
	"GhostGate/internal/logger"
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// Variables for the command
var path string
var destination string

// The upload command
var uploadCmd = &cobra.Command{
	Use: "GhostGate",
	Long: "GhostGate is a versatile Go-based networking toolkit designed for penetration testing, red teaming, and security auditing. It simplifies the process of setting up payload staging environments, running data exfiltration handlers, establishing reverse/pivot tunnels, and conducting quick HTTP security configuration audits.",

	Run: func (cmd *cobra.Command, args []string) {
		// Creates a new context
		ctx := context.Background()

		// Logs the commands are being parsed
		logger.Info(ctx, "Parsing commands for 'upload'")
		
		// Logs the ports are being cleaned
		logger.Info(ctx, "Cleaning entered port", )
		
		// Cleans the port
		cleanPort := input.CleanPort(port)

		// Logs the validation has started
		logger.Info(ctx, "Starting validation on cleaned port", cleanPort)

		// Validating the clean port
		err := input.ValidatePort(cleanPort)
		if err != nil {
			// Logs the port is invalid
			logger.Error(ctx, "Validation failed on port", cleanPort)
			
			// Prints the port is invalid
			fmt.Printf("[!] Invalid port: %s", port)
		}

		// Cleans the uploaded file path
		cleanURL := input.CleanURL(path)

		// Logs validation has started
		logger.Info(ctx, "Validation has started on path", path)

		// Validates the url
		err = input.ValidateURL(cleanURL)
		if err != nil {
			// Logs the url is invalid
			logger.Error(ctx, "Upload path is invalid", path)

			// Prints the path is invalid
			fmt.Printf("[!] Invalid upload path: %v", err)
		}

		// Logs the destination path is being cleaned
		logger.Info(ctx, "Cleaning destination file path", destination)

		// Cleans the destination path
		cleanDest := input.CleanFilePath(destination)

		// Validates the file path
		err = input.ValidateFilePath(cleanDest)
		if err != nil {
			// Logs and outputs the error
			logger.Info(ctx, "Validation failed on file path", cleanDest)

			fmt.Printf("[!] Invalid destination path: %v", err)
		}

		commands.StartUploadServer(cleanPort, cleanURL, cleanDest, useTLS, certFile, keyFile)
	},
}

// Stores the commands which are used by the program
func init()  {
	uploadCmd.Flags().StringVarP(&path, "url-path", "u", "", "Specifies the URL path to send the data to for exfilration")
	uploadCmd.Flags().StringVarP(&destination, "destination", "d", "", "Specifies the folder to store the retreived files")

	// Adds the command to the root command
	rootCmd.AddCommand(uploadCmd)
}