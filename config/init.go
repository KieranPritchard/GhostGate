package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// InitializeConfig creates the config directory and default file
func InitializeConfig() error {
	// 1. Determine path (~/.config/mytool)
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not find config directory: %w", err)
	}
	
	toolDir := filepath.Join(configDir, "GhostGate")
	configPath := filepath.Join(toolDir, "config.yaml")

	// 2. Create directory if it doesn't exist (chmod 0755 is standard)
	if _, err := os.Stat(toolDir); os.IsNotExist(err) {
		fmt.Printf("Creating config directory at %s...\n", toolDir)
		if err := os.MkdirAll(toolDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// 3. Check if file exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	// 4. Set defaults and write file
	viper.Set("api_key", "YOUR_API_KEY_HERE")
	viper.Set("verbose", false)
	
	// SafeWriteConfig ensures we don't overwrite an existing file
	if err := viper.SafeWriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Initialized default config at %s\n", configPath)
	return nil
}