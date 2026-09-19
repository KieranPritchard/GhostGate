package cmd

import (
	"GhostGate/internal/commands"
	"GhostGate/internal/input"
	"GhostGate/internal/logger"
	"GhostGate/internal/util"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// Stores the arguements relevant to the function
var headers string
var randomised bool

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

		// Creates a new request
		req, err := http.NewRequest(http.MethodGet, target.String(), nil)
		if err != nil {
			logger.Error(ctx, "HTTP GET request could not be created for target:", target)
			fmt.Println("[!] Request creation error encountered", err)
			os.Exit(1)
		}

		// Checks if randomised is true
		if randomised {
			// Gets the random the 
			randomAgent, err := util.GetRandomHeader()
			if err != nil {
				logger.Error(ctx, "Random header for target request could not be made:", target)
				fmt.Println("[!] Request creation error encountered", err)
				os.Exit(1)
			}

			req.Header.Set("User-Agent", randomAgent)
		}

		// Checks if there are any headers
		if len(headers) > 0 {
			// Prepares the headers
			structuredHeaders, err := input.PrepareHeaders(headers)
			if err != nil {
				logger.Error(ctx, "Headers could not be parsed:", headers)
				fmt.Println("[!] Header parsing error encountered:", err)
				os.Exit(1)
			}

			// Loops over each of the headers and sets them
			for _, header := range structuredHeaders {
				req.Header.Set(header.Key, header.Value)
			}
		}

		// Creates a http client
		client := &http.Client{
			// Sets timeout to 10 seconds
			Timeout: 10 * time.Second,
		}

		// Gets the response and closes the body when done
		resp, err := client.Do(req)
		if err != nil {
			logger.Error(ctx, "Connection failed", err)
			fmt.Printf("[!] Connection failed: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		commands.AuditRequest(os.Stdout, resp)
	},
}

func init(){
	// Adds the commands exclusive to this
	auditCmd.Flags().StringVarP(&headers, "headers", "-H", "", "Defines the headers which can be used in the request")
	auditCmd.Flags().BoolVarP(&randomised, "random", "r", false, "Used to randomise the user agent header")

	// Adds the root command to the audit
	rootCmd.AddCommand(auditCmd)
}