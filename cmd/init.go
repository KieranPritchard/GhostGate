package cmd

import (
	"GhostGate/config"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the GhostGate configuration file",
	Long:  `Generates a JSON configuration file for GhostGate utilizing either interactive prompts or system defaults.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := config.InitializeConfig()
		if err != nil {
			fmt.Printf("[!] Initialization failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
