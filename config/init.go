package config

import (
	"fmt"
	"os"
	"path/filepath"
	"bufio"
	"strings"

	"github.com/spf13/viper"
)

// InitializeConfig creates the config directory and default file
func InitializeConfig() error {
	configDir, _ := os.UserConfigDir()
	toolDir := filepath.Join(configDir, "GhostGate")
	configPath := filepath.Join(toolDir, "config.json")

	// 1. Create directory
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		return err
	}

	// 2. Check if file exists to prevent accidental overwrites
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config already exists at %s. Overwrite? (y/N): ", configPath)
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			return fmt.Errorf("initialization cancelled")
		}
	}

	// 3. Interactive Prompts
	reader := bufio.NewReader(os.Stdin)
	
	// Allows the user to enter a port number
	fmt.Print("Enter your port number: ")
	portNumber, _ := reader.ReadString('\n')
	portNumber = strings.TrimSpace(portNumber)

	// Checks if there wasnt anything entered
	if portNumber == "" {
        portNumber = "8080"
    }

	// Allows the user to enter a payloads directory
	fmt.Print("Enter your payloads directory: ")
	payloadDir, _ := reader.ReadString('\n')
	payloadDir = strings.TrimSpace(payloadDir)

	if payloadDir == "" {
        payloadDir = "payloads"
    }

	fmt.Print("Enter your upload path: ")
	uploadPath, _ := reader.ReadString('\n')
	uploadPath = strings.TrimSpace(uploadPath)

	if uploadPath == "" {
        uploadPath = "/uploads"
    }

	// 4. Save to Viper
	viper.Set("default_port", portNumber)
	viper.Set("default_payloads_path", payloadDir) // Was set to uploadPath in your snippet
	viper.Set("default_uploads_path", uploadPath)

	// 5. Write to file
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("\nSuccess! Configuration saved to %s\n", configPath)
	return nil
}