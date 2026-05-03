package stevedore

import (
	"github.com/jameswoolfenden/stevedore/internal/auth"
	"github.com/jameswoolfenden/stevedore/internal/dockerfile"
	"github.com/jameswoolfenden/stevedore/internal/git"
	"github.com/rs/zerolog/log"
)

// Parser is a backward compatibility wrapper
type Parser struct {
	File      *string
	Output    string
	Directory string
	Author    string
}

// ParseAll processes Dockerfiles - backward compatibility wrapper
func (content *Parser) ParseAll() error {
	authService := auth.NewDockerAuth()

	var gitService git.Service
	workDir := content.Directory
	if content.File != nil && *content.File != "" {
		workDir = *content.File
	}

	gitService, err := git.NewGitService(workDir)
	if err != nil {
		log.Warn().Err(err).Msg("git service unavailable")
		gitService = nil
	}

	labeler := dockerfile.NewLabeler(gitService, authService)
	parser := dockerfile.NewParser(labeler)

	if content.File != nil {
		parser.File = *content.File
	}
	parser.Directory = content.Directory
	parser.Output = content.Output
	parser.Author = content.Author

	return parser.ParseAll()
}

// Parse processes a single Dockerfile - backward compatibility wrapper
func (content *Parser) Parse() error {
	return content.ParseAll()
}
