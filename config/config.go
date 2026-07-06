package config

import (
	"os"
	"path/filepath"
	"github.com/spf13/viper"
)

// Defines the config struct
type Config struct {
	DefaultPort              string `mapstructure:"default_port"`
	DefaultPayloadsDirectory string `mapstructure:"default_payloads_directory"`
	DefaultUploadsDirectory  string `mapstructure:"default_uploads_directory"`
	DefaultURLPath           string `mapstructure:"default_url_path"`
	DefaultTLSEnabled        bool   `mapstructure:"default_tls_enabled"`
	DefaultTLSCertFile       string `mapstructure:"default_tls_cert_file"`
	DefaultTLSKeyFile        string `mapstructure:"default_tls_key_file"`
}

// Function to load a config file
func LoadConfig() (*Config, error) {
	// Gets the default config directory
	configDir, _ := os.UserConfigDir()

	// Creates the full file path
	fullPath := filepath.Join(configDir, "GhostGate")

	// Sets up the basic config
	viper.SetConfigName("config") // Name of the config file
	viper.SetConfigType("json")   // Type of config file
	viper.AddConfigPath(fullPath) // Stores the path where config is stored
	viper.AddConfigPath(".")      // Also checks the main directory

	// Sets the defaults
	viper.SetDefault("default_port", "8080")
	viper.SetDefault("default_payloads_directory", "payloads")
	viper.SetDefault("default_uploads_directory", "uploads")
	viper.SetDefault("default_url_path", "/uploads")
	viper.SetDefault("default_tls_enabled", false)
	viper.SetDefault("default_tls_cert_file", "")
	viper.SetDefault("default_tls_key_file", "")

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