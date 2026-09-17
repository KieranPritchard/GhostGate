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

// Defines the variables for the subcommand
var directory string
var source string

var stageCmd = &cobra.Command{
	Use: "stage",
	Short: "Serves a local directory via HTTP for remote access.",

	// Run handles the logic for when this command is called
	Run: func (cmd *cobra.Command, args []string)  {
		// Creates a new context
		ctx := context.Background()

		// Logs the commands are being parsed
		logger.Info(ctx, "Parsing commands for 'stage'")

		// Logs the ports are being cleaned
		logger.Info(ctx, "Parses the entered port", port)

		// Prepares the port
		port, err := input.PreparePort(port)
		if err != nil {
			logger.Error(ctx, "Cleaning and validation failed on target port (could not parse)", port)
			fmt.Println("[!] Port parsing error encountered:", err)
			os.Exit(1)
		}

		// Fallback to default staging directory if empty
		targetDir := directory
		if targetDir == "" {
			if cfg != nil {
				targetDir = cfg.DefaultPayloadsDirectory
			} else {
				targetDir = "payloads"
			}
		}

		logger.Info(ctx, "Parsing path for stage directory", targetDir)

		// Prepares the staging directory
		stageDir, err := input.PrepareFilePath(targetDir)
		if err != nil {
			logger.Error(ctx, "Cleaning and validation failed on stage directory (could not parse)", targetDir)
			fmt.Println("[!] Staging directory parsing error encountered:", err)
			os.Exit(1)
		}

		// Prepares the source directory
		sourceDir, err := input.PrepareFilePath(source)
		if err != nil {
			logger.Error(ctx, "Cleaning and validation failed on source directory (could not parse)", targetDir)
			fmt.Println("[!] Source directory parsing error encountered:", err)
			os.Exit(1)
		}

		// Runs the stage payload directory function
		commands.StagePayloadDirectory(port, stageDir, sourceDir, useTLS, certFile, keyFile)
	},
}

// Stores and uses the flags for this subcommand
func init() {
	// Specifies the flags
	stageCmd.Flags().StringVarP(&directory, "directory", "d", "", "Directory to host folder from")
	stageCmd.Flags().StringVarP(&source, "source", "s", "", "Directory to get the hosted files from")

	// Adds the command to the root comamand
	rootCmd.AddCommand(stageCmd)
}