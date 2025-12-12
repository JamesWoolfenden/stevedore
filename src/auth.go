package stevedore

import (
	"github.com/jameswoolfenden/stevedore/internal/auth"
)

// GetAuthToken is a backward compatibility wrapper for Docker authentication
func GetAuthToken(from string) (string, error) {
	authService := auth.NewDockerAuth()
	return authService.GetAuthToken(from)
}
