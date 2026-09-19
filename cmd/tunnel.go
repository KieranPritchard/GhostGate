package cmd

import (
	"GhostGate/internal/commands"
	"GhostGate/internal/input"
	"GhostGate/internal/logger"
	"GhostGate/internal/util"
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
			fmt.Println("[!] Domain parsing error encountered: ", err)
			os.Exit(1)
		}

		// Logs the ports are being cleaned
		logger.Info(ctx, "Parses the entered port", port)

		// Prepares the port
		port, err := input.PreparePort(port)
		if err != nil {
			logger.Error(ctx, "Cleaning and validation failed on target port (could not parse)", port)
			fmt.Println("[!] Port parsing error encountered:", err)
			os.Exit(1)
		}

		var randomAgent string
		var structuredHeaders []input.Header

		// Checks if randomised is true
		if randomised {
			// Gets the random the 
			randomAgent, err = util.GetRandomHeader()
			if err != nil {
				logger.Error(ctx, "Random header for target request could not be made:", target)
				fmt.Println("[!] Request creation error encountered", err)
				os.Exit(1)
			}
		}

		// Checks if there are any headers
		if len(headers) > 0 {
			// Prepares the headers
			structuredHeaders, err = input.PrepareHeaders(headers)
			if err != nil {
				logger.Error(ctx, "Headers could not be parsed:", headers)
				fmt.Println("[!] Header parsing error encountered:", err)
				os.Exit(1)
			}
		}

		commands.StartTunnelServer(port, target.String(), useTLS, certFile, keyFile, randomAgent, structuredHeaders)
	},
}

// Stores the commands which are used by the program
func init() {
	// Adds the command to the root file
	rootCmd.AddCommand(tunnelCmd)
}