package config

import (
	"os"
	"path/filepath"
	"github.com/spf13/viper"
)

// Defines the config struct
type Config struct {
	DefaultPort string
	DefaultPayloadsDirectory string
	DefaultURLPath string
}

// Function to load a config file
func LoadConfig() (*Config, error) {
	// Gets the default config directory
	configDir, _ := os.UserConfigDir()

	// Creates the full file path
	fullPath := filepath.Join(configDir, "GhostGate")

	// Sets up the basic config
	viper.SetConfigName("config") // Name of the config file
	viper.SetConfigType("json") // Type of config file
	viper.AddConfigPath(fullPath) // Stores the path where config is stored
	viper.AddConfigPath(".") // Also checks the main directory

	// Sets the defaults
	viper.SetDefault("default_port", "8080")
	viper.SetDefault("default_payloads_directory", "payloads")
	viper.SetDefault("default_url_path", "/uploads")

	// Sets the enviroment variables
	// ADD HERE WHEN NEEDED

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	err := viper.Unmarshal(&cfg)
	return &cfg, err
}