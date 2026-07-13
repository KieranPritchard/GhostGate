package cmd

import (
	"GhostGate/config"
	"GhostGate/internal/logger"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Stores the config
var cfg *config.Config

// Defines the global commands
var useTLS bool
var certFile string
var keyFile string
var port string
var target string

// Defines the root command
var rootCmd = &cobra.Command{
	Use:  "GhostGate",
	Long: "GhostGate is a versatile Go-based networking toolkit designed for penetration testing, red teaming, and security auditing. It simplifies the process of setting up payload staging environments, running data exfiltration handlers, establishing reverse/pivot tunnels, and conducting quick HTTP security configuration audits.",

	// Creates a new logger before running each subcommand
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Creates a new logger
		logger.New(logger.Config{
			Level:  "INFO",
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
	// Loads the config
	loadedCfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}
	cfg = loadedCfg

	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", cfg.DefaultPort, "Port to run the service on")
	rootCmd.PersistentFlags().BoolVarP(&useTLS, "tls", "e", cfg.DefaultTLSEnabled, "Specifies to use tls for connection")
	rootCmd.PersistentFlags().StringVarP(&certFile, "cert-file", "c", cfg.DefaultTLSCertFile, "Specifies a path of a cert file")
	rootCmd.PersistentFlags().StringVarP(&keyFile, "key-file", "k", cfg.DefaultTLSKeyFile, "Specifies a path of a key file")
	rootCmd.PersistentFlags().StringVarP(&target, "target", "t", "", "Specifies the target")
}