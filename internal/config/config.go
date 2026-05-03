package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds the application configuration
type Config struct {
	DefaultAuthor string
	Output        string
	LogLevel      string
	HTTPTimeout   time.Duration
}

// NewConfig creates a new configuration with sensible defaults
func NewConfig() *Config {
	return &Config{
		DefaultAuthor: "",
		Output:        ".",
		LogLevel:      "info",
		HTTPTimeout:   30 * time.Second,
	}
}

// Validate performs validation on the configuration
func (c *Config) Validate() error {
	if c.Output != "" {
		if err := validatePath(c.Output); err != nil {
			return fmt.Errorf("invalid output path: %w", err)
		}
	}

	return nil
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

// validatePath performs security validation on file paths to prevent path traversal
func validatePath(path string) error {
	if path == "" {
		return nil // Empty path is handled elsewhere
	}

	cleaned := filepath.Clean(path)

	// Check for path traversal attempts
	if containsPathTraversal(cleaned) {
		return fmt.Errorf("path traversal detected in: %s", path)
	}

	return nil
}

// containsPathTraversal checks if a path contains path traversal patterns
func containsPathTraversal(path string) bool {
	// Normalize the path
	cleaned := filepath.Clean(path)

	// Check if the cleaned path goes outside of current directory
	// by checking for leading ".."
	parts := filepath.SplitList(cleaned)
	for _, part := range parts {
		if part == ".." {
			return true
		}
	}

	return false
}

// ValidateDockerfilePath validates a Dockerfile path for security
func ValidateDockerfilePath(path string) error {
	if path == "" {
		return fmt.Errorf("dockerfile path cannot be empty")
	}

	return validatePath(path)
}
