package cmd

import (
	"GhostGate/internal/commands"
	"GhostGate/internal/input"
	"GhostGate/internal/logger"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use: "audit",
	Short: "Audits a http server",

	Run: func (cmd *cobra.Command, args []string) {
		// Creates a new context
		ctx := context.Background()

		// Logs the commands are being parsed
		logger.Info(ctx, "Parsing commands for 'audit'")

		// Cleans the url
		cleanTarget := input.CleanURL(target)

		// Logs the url validation has started
		logger.Info(ctx, "Starting validation on the url")

		// Validates the target
		err := input.ValidateURL(cleanTarget)
		
		// Checks if the url is valid
		if err != nil {
			// Logs the url is incorrect
			logger.Error(ctx, "Invalid url", target)

			// Outputs the url is incorrect
			fmt.Printf("[!] Invalid audit target URL: %v\n", err)
		}

		// Prints the configuration is starting
		fmt.Printf("[*] Launching configuration audit against: %s\n", target)

		// Creates a http client
		client := &http.Client{
			// Sets timeout to 10 seconds
			Timeout: 10 * time.Second,
		}

		// Gets the response
		resp, err := client.Get(target)
		if err != nil {
			// Logs the response failed
			logger.Error(ctx, "Connection failed", err)
			fmt.Printf("[!] Connection failed: %v\n", err)
		}
		defer resp.Body.Close()

		commands.AuditRequest(resp)
	},
}