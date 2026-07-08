package cmd

import (
	"GhostGate/internal/commands"
	"GhostGate/internal/input"
	"GhostGate/internal/logger"
	"context"
	"fmt"

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
		logger.Info(ctx, "Cleaning entered port", port)

		// Cleans the port number
		cleanPort := input.CleanPort(port)

		// Logs the validation has started
		logger.Info(ctx, "Starting validation on cleaned port", cleanPort)

		// Validates the port number
		err := input.ValidatePort(cleanPort)

		// Validates the port number
		if err != nil{
			// Logs the port is invalid
			logger.Error(ctx, "Validation failed on port", cleanPort)

			// Outputs the port is invalid
			fmt.Printf("[!] Invalid port: %s\n", port)
		}

		// Logs the file path is being cleaned
		logger.Info(ctx, "Cleaning path for stage directory", port)
		
		// Cleans the path for the staging directory
		cleanDir := input.CleanFilePath(directory)

		// Logs the directory is being cleaned
		logger.Info(ctx, "Validating the clean staging directory", cleanDir)

		// Checks if the clean directory is valid
		err = input.ValidateFilePath(cleanDir)
		
		if err != nil {
			// Logs the validation has failed
			logger.Info(ctx, "Validation of the staging directory has failed", cleanDir)
			
			// Outputs the staging directory is invalid
			fmt.Printf("[!] Invalid staging directory: %s", directory)
		}

		// The source flag is optional — only validate it when the user provided a value
		cleanSource := ""

		// Checks if a stage source was entered
		if source != "" {
			
			// Logs if the source directory is being validated
			logger.Info(ctx, "Validating source directory for the staging", source)

			// Validates the source path
			err = input.ValidateFilePath(input.CleanFilePath(source))
			if err != nil {
				// logs the source path is invalid
				logger.Error(ctx, "Invalid source directory", source)

				// Outputs the source is invalid
				fmt.Printf("[!] Invalid source directory: %s", source)
			}
		}

		// Runs the stage payload directory function
		commands.StagePayloadDirectory(port, cleanDir, cleanSource, useTLS, certFile, keyFile)
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