package cmd

import (
	"GhostGate/internal/commands"
	"GhostGate/internal/input"
	"GhostGate/internal/logger"
	"context"
	"fmt"
	"net/http"
	"os"
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

		// Prepares the target
		target, err := input.PrepareURL(target)
		if err != nil {
			logger.Error(ctx, "Cleaning and validation failed on target URL (could not parse)", target)
			fmt.Println("[!] Auditing error encountered", err)
			os.Exit(1)
		}

		fmt.Printf("[*] Launching configuration audit against: %s\n", target.String())

		// Creates a http client
		client := &http.Client{
			// Sets timeout to 10 seconds
			Timeout: 10 * time.Second,
		}

		// Gets the response
		resp, err := client.Get(target.String())
		if err != nil {
			// Logs the response failed
			logger.Error(ctx, "Connection failed", err)
			fmt.Printf("[!] Connection failed: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		commands.AuditRequest(os.Stdout, resp)
	},
}

func init(){
	rootCmd.AddCommand(auditCmd)
}