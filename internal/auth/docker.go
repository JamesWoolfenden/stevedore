package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// DockerAuth defines the interface for Docker registry authentication
type DockerAuth interface {
	GetAuthToken(image string) (string, error)
}

// dockerAuthService implements Docker registry authentication
type dockerAuthService struct {
	client *http.Client
}

// NewDockerAuth creates a new Docker authentication service with proper HTTP timeouts
func NewDockerAuth() DockerAuth {
	return &dockerAuthService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAuthToken retrieves an authentication token from Docker Hub for the specified image
func (d *dockerAuthService) GetAuthToken(image string) (string, error) {
	if err := validateImageName(image); err != nil {
		return "", err
	}

	url := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + image + ":pull"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute http request: %w", err)
	}

	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close http response body")
		}
	}()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(body, &jsonMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	token, ok := jsonMap["token"].(string)
	if !ok {
		return "", fmt.Errorf("token not found in response or invalid type")
	}

	if token == "" {
		return "", fmt.Errorf("empty token received")
	}

	return token, nil
}

// validateImageName performs basic validation on Docker image names
func validateImageName(image string) error {
	if image == "" {
		return fmt.Errorf("image name cannot be empty")
	}
	// Add more validation as needed
	return nil
}
