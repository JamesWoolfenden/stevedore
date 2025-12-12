package dockerfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jameswoolfenden/stevedore/internal/config"
	"github.com/rs/zerolog/log"
)

// Parser coordinates the scanning and processing of Dockerfiles
type Parser struct {
	File      string
	Output    string
	Directory string
	Author    string
	labeller   *Labeller
}

// NewParser creates a new Parser instance
func NewParser(labeller *Labeller) *Parser {
	return &Parser{
		labeller: labeller,
		Output:  ".",
	}
}

// ParseAll processes either a single file or all Dockerfiles in a directory
func (p *Parser) ParseAll() error {
	if p.File != "" {
		return p.parseSingleFile()
	}

	return p.parseDirectory()
}

// parseSingleFile processes a single Dockerfile
func (p *Parser) parseSingleFile() error {
	if err := config.ValidateDockerfilePath(p.File); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	return p.parseFile(p.File)
}

// parseDirectory walks a directory tree and processes all Dockerfiles
func (p *Parser) parseDirectory() error {
	if p.Directory == "" {
		p.Directory = "."
	}

	if err := config.ValidateDockerfilePath(p.Directory); err != nil {
		return fmt.Errorf("invalid directory path: %w", err)
	}

	err := filepath.Walk(p.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.Contains(info.Name(), "Dockerfile") {
			if parseErr := p.parseFile(path); parseErr != nil {
				log.Error().Err(parseErr).Msgf("failed to parse %s", path)
				return parseErr
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("directory walk failed: %w", err)
	}

	return nil
}

// parseFile parses a single Dockerfile and writes the labeled version
func (p *Parser) parseFile(filePath string) error {
	dockerfile := &Dockerfile{
		Path: filePath,
	}

	if err := dockerfile.ParseFile(); err != nil {
		return fmt.Errorf("failed to parse dockerfile: %w", err)
	}

	dump, err := p.labeller.Label(dockerfile, p.Author)
	if err != nil {
		return fmt.Errorf("failed to add labels: %w", err)
	}

	outputPath := filepath.Join(p.Output, filepath.Base(filePath))

	//#nosec
	if err := os.WriteFile(outputPath, []byte(dump), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	log.Info().Msgf("updated: %s", filepath.Base(filePath))

	return nil
}
