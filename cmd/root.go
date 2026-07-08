package cmd

import (
	"GhostGate/internal/logger"
	"os"

	"github.com/spf13/cobra"
)

// Defines the global commands
var useTLS bool
var certFile string
var keyFile string
var port string

// Defines the root command
var rootCmd = &cobra.Command{
	Use: "GhostGate",
	Long: "GhostGate is a versatile Go-based networking toolkit designed for penetration testing, red teaming, and security auditing. It simplifies the process of setting up payload staging environments, running data exfiltration handlers, establishing reverse/pivot tunnels, and conducting quick HTTP security configuration audits.",

	// Creates a new logger before running each subcommand
	PersistentPreRun: func (cmd *cobra.Command, args []string)  {
		// Creates a new logger
		logger.New(logger.Config{
			Level: "INFO",
			Format: logger.FormatText,
		})
	},
}

// Function to execute the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// Add global subcommands under init function
func init() {
	stageCmd.Flags().StringVarP(&port, "port", "p", "", "Port to run the service on")
	rootCmd.Flags().BoolVarP(&useTLS, "tls", "tls", false, "Specifies to use tls for connection")
	rootCmd.Flags().StringVarP(&certFile, "cert-file", "c", "", "Specifies a path of a cert file")
	rootCmd.Flags().StringVarP(&certFile, "key-file", "k", "", "Specifies a path of a key file")
}