package cmd

import (
	"GhostGate/internal/commands"
	"GhostGate/internal/input"
	"GhostGate/internal/logger"
	"context"
	"fmt"
	"os"

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

		logger.Info(ctx, "Parsing path for target", targetPath)

		// Prepares the staging directory
		targetURLPath, err := input.PrepareFilePath(targetPath)
		if err != nil {
			logger.Error(ctx, "Cleaning and validation failed on target (could not parse)", targetURLPath)
			fmt.Println("[!] Target parsing error encountered:", err)
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

		logger.Info(ctx, "Parsing path for target", targetPath)

		// Prepares the staging directory
		targetDestPath, err := input.PrepareFilePath(targetPath)
		if err != nil {
			logger.Error(ctx, "Cleaning and validation failed on target destination (could not parse)", targetDestPath)
			fmt.Println("[!] Target destination parsing error encountered:", err)
			os.Exit(1)
		}

		commands.StartUploadServer(port, targetURLPath, targetDest, useTLS, certFile, keyFile)
	},
}

// Stores the commands which are used by the program
func init()  {
	uploadCmd.Flags().StringVarP(&path, "url", "u", "", "Specifies the URL path to send the data to for exfilration")
	uploadCmd.Flags().StringVarP(&destination, "destination", "d", "", "Specifies the folder to store the retreived files")

	// Adds the command to the root command
	rootCmd.AddCommand(uploadCmd)
}