package cmd

import (
	"GhostGate/internal/commands"
	"GhostGate/internal/input"
	"GhostGate/internal/logger"
	"context"
	"fmt"
	"os"
	"strings"

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
		logger.Info(ctx, "Parses the entered port", port)

		// Prepares the port
		port, err := input.PreparePort(port)
		if err != nil {
			logger.Error(ctx, "Cleaning and validation failed on target port (could not parse)", port)
			fmt.Println("[!] Port parsing error encountered:", err)
			os.Exit(1)
		}

		// Fallback to default url path if empty
		targetPath := path
		if targetPath == "" {
			if cfg != nil {
				targetPath = cfg.DefaultURLPath
			} else {
				targetPath = "/uploads"
			}
		}

		// Cleans the uploaded URL path
		cleanPath := strings.TrimSpace(targetPath)
		if !strings.HasPrefix(cleanPath, "/") {
			cleanPath = "/" + cleanPath
		}

		// Logs validation has started
		logger.Info(ctx, "Validation has started on path", cleanPath)

		// Validates the url path
		if cleanPath == "/" || strings.ContainsAny(cleanPath, " ?#") {
			// Logs the path is invalid
			logger.Error(ctx, "Upload path is invalid", cleanPath)

			// Prints the path is invalid
			fmt.Printf("[!] Invalid upload path: %s\n", cleanPath)
			os.Exit(1)
		}

		// Fallback to default uploads directory if empty
		targetDest := destination
		if targetDest == "" {
			if cfg != nil {
				targetDest = cfg.DefaultUploadsDirectory
			} else {
				targetDest = "uploads"
			}
		}

		// Logs the destination path is being cleaned
		logger.Info(ctx, "Cleaning destination file path", targetDest)

		// Cleans the destination path
		cleanDest := input.CleanFilePath(targetDest)

		// Validates the file path
		err = input.ValidateFilePath(cleanDest)
		if err != nil {
			// Logs and outputs the error
			logger.Info(ctx, "Validation failed on file path", cleanDest)

			fmt.Printf("[!] Invalid destination path: %v\n", err)
			os.Exit(1)
		}

		commands.StartUploadServer(port, cleanPath, cleanDest, useTLS, certFile, keyFile)
	},
}

// Stores the commands which are used by the program
func init()  {
	uploadCmd.Flags().StringVarP(&path, "url-path", "u", "", "Specifies the URL path to send the data to for exfilration")
	uploadCmd.Flags().StringVarP(&destination, "destination", "d", "", "Specifies the folder to store the retreived files")

	// Adds the command to the root command
	rootCmd.AddCommand(uploadCmd)
}