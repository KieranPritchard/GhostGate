package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// InitializeConfig creates the config directory and writes an interactive default config file.
func InitializeConfig() error {
	configDir, _ := os.UserConfigDir()
	toolDir := filepath.Join(configDir, "GhostGate")
	configPath := filepath.Join(toolDir, "config.json")

	// Create the config directory if it doesn't exist
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		return err
	}

	// Check if a config file already exists to prevent accidental overwrites
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config already exists at %s. Overwrite? (y/N): ", configPath)
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			return fmt.Errorf("initialization cancelled")
		}
	}

	// Interactive prompts
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your default port number [8080]: ")
	portNumber, _ := reader.ReadString('\n')
	portNumber = strings.TrimSpace(portNumber)
	if portNumber == "" {
		portNumber = "8080"
	}

	fmt.Print("Enter your payloads directory [payloads]: ")
	payloadDir, _ := reader.ReadString('\n')
	payloadDir = strings.TrimSpace(payloadDir)
	if payloadDir == "" {
		payloadDir = "payloads"
	}

	fmt.Print("Enter your default uploads directory [uploads]: ")
	uploadDir, _ := reader.ReadString('\n')
	uploadDir = strings.TrimSpace(uploadDir)
	if uploadDir == "" {
		uploadDir = "uploads"
	}

	fmt.Print("Enter your upload URL path [/uploads]: ")
	uploadPath, _ := reader.ReadString('\n')
	uploadPath = strings.TrimSpace(uploadPath)
	if uploadPath == "" {
		uploadPath = "/uploads"
	}

	fmt.Print("Enable TLS by default? (y/N): ")
	tlsInput, _ := reader.ReadString('\n')
	tlsEnabled := strings.ToLower(strings.TrimSpace(tlsInput)) == "y"

	var certFile, keyFile string
	if tlsEnabled {
		fmt.Print("Enter path to default TLS certificate file (leave empty for auto-generated): ")
		certFile, _ = reader.ReadString('\n')
		certFile = strings.TrimSpace(certFile)

		if certFile != "" {
			fmt.Print("Enter path to default TLS private key file: ")
			keyFile, _ = reader.ReadString('\n')
			keyFile = strings.TrimSpace(keyFile)
		}
	}

	// Save values using the same keys defined in config.go and its struct tags
	viper.Set("default_port", portNumber)
	viper.Set("default_payloads_directory", payloadDir)
	viper.Set("default_uploads_directory", uploadDir)
	viper.Set("default_url_path", uploadPath)
	viper.Set("default_tls_enabled", tlsEnabled)
	viper.Set("default_tls_cert_file", certFile)
	viper.Set("default_tls_key_file", keyFile)

	// Write to disk
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("\nSuccess! Configuration saved to %s\n", configPath)
	return nil
}