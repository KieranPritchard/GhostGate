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