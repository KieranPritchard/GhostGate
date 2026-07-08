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
	Use: "upload",
	Short: "Creates a URL which allows for the user upload files to",

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
			fmt.Printf("[!] Invalid port: %s\n", port)
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
			fmt.Printf("[!] Invalid upload path: %v\n", err)
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

			fmt.Printf("[!] Invalid destination path: %v\n", err)
		}

		commands.StartUploadServer(cleanPort, cleanURL.String(), cleanDest, useTLS, certFile, keyFile)
	},
}

// Stores the commands which are used by the program
func init()  {
	uploadCmd.Flags().StringVarP(&path, "url-path", "u", "", "Specifies the URL path to send the data to for exfilration")
	uploadCmd.Flags().StringVarP(&destination, "destination", "d", "", "Specifies the folder to store the retreived files")

	// Adds the command to the root command
	rootCmd.AddCommand(uploadCmd)
}