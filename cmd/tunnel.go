package cmd

import (
	"GhostGate/internal/commands"
	"GhostGate/internal/input"
	"GhostGate/internal/logger"
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// The tunnel command
var tunnelCmd = &cobra.Command{
	Use: "tunnel",
	Short: "Creates a http tunnel to a target",

	Run: func (cmd *cobra.Command, args []string)  {
		// Creates a new context
		ctx := context.Background()

		// Logs the commands are being parsed
		logger.Info(ctx, "Parsing commands for 'tunnel'")

		// Cleans the tunnel target
		cleanURL := input.CleanURL(target)

		// Logs validation has start
		logger.Info(ctx, "Validation started on the url", cleanURL)

		// Validates the url
		err := input.ValidateURL(cleanURL)

		// Check if there is an error
		if err != nil {
			// Logs the validation has failed
			logger.Error(ctx, "Validation failed on tunnel target", target)
			
			// Outputs the target is invalid
			fmt.Printf("[!] Invalid tunnel target URL: %v\n", err)
		}

		// Logs the port is being cleaned
		logger.Info(ctx, "Cleaning has started on the port", port)

		// Cleans the port entered
		cleanPort := input.CleanPort(port)

		// Validates the path
		err = input.ValidatePort(cleanPort)
		if err != nil {
			// Logs the port is invalid
			logger.Error(ctx, "Port is invalid", port)

			// Ouputs the port is invalid
			fmt.Printf("[!] Invalid port: %s\n", port)
		}

		commands.StartTunnelServer(cleanPort, target, useTLS, certFile, keyFile)
	},
}

// Stores the commands which are used by the program
func init() {
	tunnelCmd.Flags().StringVarP(&target, "target", "t", "", "Specifies the target to tunnel to")

	// Adds the command to the root file
	rootCmd.AddCommand(tunnelCmd)
}