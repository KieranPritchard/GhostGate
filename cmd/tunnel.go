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

// The tunnel command
var tunnelCmd = &cobra.Command{
	Use: "tunnel",
	Short: "Creates a http tunnel to a target",

	Run: func (cmd *cobra.Command, args []string)  {
		// Creates a new context
		ctx := context.Background()

		// Logs the commands are being parsed
		logger.Info(ctx, "Parsing commands for 'tunnel'")

		// Prepares the target
		target, err := input.PrepareURL(target)
		if err != nil {
			logger.Error(ctx, "Cleaning and validation failed on target (could not parse)", target)
			fmt.Println("[!] Auditing error encountered", err)
			os.Exit(1)
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
			os.Exit(1)
		}

		commands.StartTunnelServer(cleanPort, target.String(), useTLS, certFile, keyFile)
	},
}

// Stores the commands which are used by the program
func init() {
	tunnelCmd.Flags().StringVarP(&target, "target", "t", "", "Specifies the target to tunnel to")

	// Adds the command to the root file
	rootCmd.AddCommand(tunnelCmd)
}