package stevedore

import (
	"os/user"

	"github.com/jameswoolfenden/stevedore/internal/auth"
	"github.com/jameswoolfenden/stevedore/internal/dockerfile"
	"github.com/jameswoolfenden/stevedore/internal/git"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/rs/zerolog/log"
)

// Dockerfile is a backward compatibility wrapper
type Dockerfile struct {
	Parsed *parser.Result
	Path   string
	Image  string
}

// ParseFile parses a Dockerfile - backward compatibility wrapper
func (result *Dockerfile) ParseFile() error {
	df := &dockerfile.Dockerfile{
		Parsed: result.Parsed,
		Path:   result.Path,
		Image:  result.Image,
	}
	err := df.ParseFile()
	result.Parsed = df.Parsed
	return err
}

// Label adds metadata labels to the Dockerfile - backward compatibility wrapper
func (result *Dockerfile) Label(Author string) (string, error) {
	// Initialize services
	authService := auth.NewDockerAuth()

	// Try to get git service, but don't fail if unavailable
	var gitService git.Service
	gitService, err := git.NewGitService(result.Path)
	if err != nil {
		log.Warn().Err(err).Msg("git service unavailable in backward compat mode")
		gitService = nil
	}

	labeler := dockerfile.NewLabeler(gitService, authService)
	df := &dockerfile.Dockerfile{
		Parsed: result.Parsed,
		Path:   result.Path,
		Image:  result.Image,
	}

	output, err := labeler.Label(df, Author)
	result.Parsed = df.Parsed
	return output, err
}

// MakeLabel is kept for backward compatibility but is deprecated
// The new implementation is in internal/dockerfile/labeler.go
func MakeLabel(child *parser.Node, layer int64, myUser *user.User, endLine int, file *string) *parser.Node {
	// This function is no longer used by the main code but is kept for test compatibility
	log.Warn().Msg("MakeLabel is deprecated, use internal/dockerfile.Labeler instead")
	return child
}

// GetDockerLabels retrieves labels from parent image - backward compatibility wrapper
func (result *Dockerfile) GetDockerLabels() (map[string]interface{}, error) {
	authService := auth.NewDockerAuth()

	// Try to get git service, but don't fail if unavailable
	var gitService git.Service
	gitService, _ = git.NewGitService(result.Path)

	labeler := dockerfile.NewLabeler(gitService, authService)
	df := &dockerfile.Dockerfile{
		Parsed: result.Parsed,
		Path:   result.Path,
		Image:  result.Image,
	}

	return labeler.GetDockerLabels(df)
}
