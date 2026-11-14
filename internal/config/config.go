package config

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// DefaultHTTPTimeout is the timeout for calls to the Docker registry
const DefaultHTTPTimeout = 30 * time.Second

// Config holds the application configuration
type Config struct {
	DefaultAuthor string
	LogLevel      string
}

// NewConfig creates a new configuration with sensible defaults
func NewConfig() *Config {
	return &Config{
		DefaultAuthor: "",
		LogLevel:      "info",
	}
}

// SetupLogging configures the logging based on the config
func (c *Config) SetupLogging() error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	level, err := zerolog.ParseLevel(c.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level %s: %w", c.LogLevel, err)
	}

	zerolog.SetGlobalLevel(level)

	return nil
}

// ValidateDockerfilePath checks that a Dockerfile or directory path was given.
//
// It does not reject ".." components: this is a local tool run by the person who
// owns the files, and "-f ../Dockerfile" is a normal thing to pass.
func ValidateDockerfilePath(path string) error {
	if path == "" {
		return fmt.Errorf("dockerfile path cannot be empty")
	}

	return nil
}
