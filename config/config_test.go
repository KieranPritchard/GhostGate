package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// resetViper clears viper state between tests so they don't bleed into each other.
func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
}

// TestLoadConfig_Defaults verifies that all default values are applied when no config file exists.
func TestLoadConfig_Defaults(t *testing.T) {
	resetViper(t)

	// Point viper at a directory that definitely has no config file.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir) // On macOS os.UserConfigDir() uses $HOME/Library/...
	// We rely on viper finding no file and returning defaults.

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if cfg.DefaultPort != "8080" {
		t.Errorf("DefaultPort = %q; want %q", cfg.DefaultPort, "8080")
	}
	if cfg.DefaultPayloadsDirectory != "payloads" {
		t.Errorf("DefaultPayloadsDirectory = %q; want %q", cfg.DefaultPayloadsDirectory, "payloads")
	}
	if cfg.DefaultUploadsDirectory != "uploads" {
		t.Errorf("DefaultUploadsDirectory = %q; want %q", cfg.DefaultUploadsDirectory, "uploads")
	}
	if cfg.DefaultURLPath != "/uploads" {
		t.Errorf("DefaultURLPath = %q; want %q", cfg.DefaultURLPath, "/uploads")
	}
	if cfg.DefaultTLSEnabled != false {
		t.Errorf("DefaultTLSEnabled = %v; want false", cfg.DefaultTLSEnabled)
	}
	if cfg.DefaultTLSCertFile != "" {
		t.Errorf("DefaultTLSCertFile = %q; want empty", cfg.DefaultTLSCertFile)
	}
	if cfg.DefaultTLSKeyFile != "" {
		t.Errorf("DefaultTLSKeyFile = %q; want empty", cfg.DefaultTLSKeyFile)
	}
}

// TestLoadConfig_FromFile verifies that values from a config file override defaults.
func TestLoadConfig_FromFile(t *testing.T) {
	resetViper(t)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Write a custom config.
	customCfg := map[string]interface{}{
		"default_port":               "9090",
		"default_payloads_directory": "custom_payloads",
		"default_uploads_directory":  "custom_uploads",
		"default_url_path":           "/custom",
		"default_tls_enabled":        true,
		"default_tls_cert_file":      "/path/to/cert.pem",
		"default_tls_key_file":       "/path/to/key.pem",
	}
	data, err := json.Marshal(customCfg)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Tell viper to load from our temp directory.
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(tmpDir)

	// Set the same defaults LoadConfig sets so viper.Unmarshal has them.
	viper.SetDefault("default_port", "8080")
	viper.SetDefault("default_payloads_directory", "payloads")
	viper.SetDefault("default_uploads_directory", "uploads")
	viper.SetDefault("default_url_path", "/uploads")
	viper.SetDefault("default_tls_enabled", false)
	viper.SetDefault("default_tls_cert_file", "")
	viper.SetDefault("default_tls_key_file", "")

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("viper.ReadInConfig() failed: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		t.Fatalf("viper.Unmarshal() failed: %v", err)
	}

	if cfg.DefaultPort != "9090" {
		t.Errorf("DefaultPort = %q; want %q", cfg.DefaultPort, "9090")
	}
	if cfg.DefaultPayloadsDirectory != "custom_payloads" {
		t.Errorf("DefaultPayloadsDirectory = %q; want %q", cfg.DefaultPayloadsDirectory, "custom_payloads")
	}
	if cfg.DefaultUploadsDirectory != "custom_uploads" {
		t.Errorf("DefaultUploadsDirectory = %q; want %q", cfg.DefaultUploadsDirectory, "custom_uploads")
	}
	if cfg.DefaultURLPath != "/custom" {
		t.Errorf("DefaultURLPath = %q; want %q", cfg.DefaultURLPath, "/custom")
	}
	if !cfg.DefaultTLSEnabled {
		t.Error("DefaultTLSEnabled = false; want true")
	}
	if cfg.DefaultTLSCertFile != "/path/to/cert.pem" {
		t.Errorf("DefaultTLSCertFile = %q; want %q", cfg.DefaultTLSCertFile, "/path/to/cert.pem")
	}
	if cfg.DefaultTLSKeyFile != "/path/to/key.pem" {
		t.Errorf("DefaultTLSKeyFile = %q; want %q", cfg.DefaultTLSKeyFile, "/path/to/key.pem")
	}
}

// TestConfig_Struct verifies the Config struct can be marshalled/unmarshalled correctly.
func TestConfig_Struct(t *testing.T) {
	original := Config{
		DefaultPort:              "1234",
		DefaultPayloadsDirectory: "pl",
		DefaultUploadsDirectory:  "ul",
		DefaultURLPath:           "/p",
		DefaultTLSEnabled:        true,
		DefaultTLSCertFile:       "cert.pem",
		DefaultTLSKeyFile:        "key.pem",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.DefaultPort != original.DefaultPort {
		t.Errorf("DefaultPort round-trip: got %q; want %q", decoded.DefaultPort, original.DefaultPort)
	}
	if decoded.DefaultTLSEnabled != original.DefaultTLSEnabled {
		t.Errorf("DefaultTLSEnabled round-trip: got %v; want %v", decoded.DefaultTLSEnabled, original.DefaultTLSEnabled)
	}
}

// TestInitializeConfig_Defaults verifies InitializeConfig creates a valid config with defaults when given empty inputs.
func TestInitializeConfig_Defaults(t *testing.T) {
	resetViper(t)
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	input := "\n\n\n\n\n" // accept all defaults, TLS=no
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
	}()

	go func() {
		_, _ = w.Write([]byte(input))
		_ = w.Close()
	}()

	err = InitializeConfig()
	if err != nil {
		t.Fatalf("InitializeConfig() failed: %v", err)
	}

	configDir, _ := os.UserConfigDir()
	configPath := filepath.Join(configDir, "GhostGate", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("expected config file at %s, but not found", configPath)
	}
}

// TestInitializeConfig_ExistingFile_Cancel verifies that choosing 'n' on overwrite aborts without modifying.
func TestInitializeConfig_ExistingFile_Cancel(t *testing.T) {
	resetViper(t)
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	configDir, _ := os.UserConfigDir()
	toolDir := filepath.Join(configDir, "GhostGate")
	_ = os.MkdirAll(toolDir, 0755)
	configPath := filepath.Join(toolDir, "config.json")
	_ = os.WriteFile(configPath, []byte("{}"), 0644)

	input := "n\n"
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
	}()

	go func() {
		_, _ = w.Write([]byte(input))
		_ = w.Close()
	}()

	err = InitializeConfig()
	if err == nil {
		t.Fatal("expected error on cancel, got nil")
	}
}

// TestInitializeConfig_WithTLS verifies InitializeConfig with TLS and custom fields.
func TestInitializeConfig_WithTLS(t *testing.T) {
	resetViper(t)
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	input := "9090\ncustom_p\ncustom_u\n/custom\ny\n/path/cert.pem\n/path/key.pem\n"
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
	}()

	go func() {
		_, _ = w.Write([]byte(input))
		_ = w.Close()
	}()

	err = InitializeConfig()
	if err != nil {
		t.Fatalf("InitializeConfig() failed: %v", err)
	}
}

