package dockerfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jameswoolfenden/stevedore/internal/config"
	"github.com/rs/zerolog/log"
)

// Parser coordinates the scanning and processing of Dockerfiles
type Parser struct {
	File string
	// Output is where updated Dockerfiles go. Empty (the default) updates each
	// Dockerfile in place; otherwise the tree under Directory is mirrored below it.
	Output    string
	Directory string
	Author    string
	labeller  *Labeller
}

// NewParser creates a new Parser instance
func NewParser(labeller *Labeller) *Parser {
	return &Parser{
		labeller: labeller,
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

	return p.parseFile(p.File, filepath.Base(p.File))
}

// skippedDirs are never searched for Dockerfiles
var skippedDirs = map[string]bool{".git": true, "node_modules": true, "vendor": true}

// notDockerfileExt are suffixes that make "Dockerfile.<ext>" a document or backup, not a Dockerfile
var notDockerfileExt = map[string]bool{
	".md": true, ".txt": true, ".bak": true, ".orig": true, ".swp": true,
	".tmp": true, ".yml": true, ".yaml": true, ".json": true,
}

// isDockerfileName reports whether a file name is Dockerfile, Dockerfile.<tag> or <name>.Dockerfile
func isDockerfileName(name string) bool {
	if name == "Dockerfile" || strings.HasSuffix(name, ".Dockerfile") {
		return true
	}

	if rest, ok := strings.CutPrefix(name, "Dockerfile."); ok {
		return !notDockerfileExt[strings.ToLower(filepath.Ext(name))] && rest != ""
	}

	return false
}

// parseDirectory walks a directory tree and processes all Dockerfiles
func (p *Parser) parseDirectory() error {
	if p.Directory == "" {
		p.Directory = "."
	}

	if err := config.ValidateDockerfilePath(p.Directory); err != nil {
		return fmt.Errorf("invalid directory path: %w", err)
	}

	var failures []error

	err := filepath.Walk(p.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if path != p.Directory && skippedDirs[info.Name()] {
				return filepath.SkipDir
			}

			return nil
		}

		if !isDockerfileName(info.Name()) {
			return nil
		}

		rel, relErr := filepath.Rel(p.Directory, path)
		if relErr != nil {
			rel = info.Name()
		}

		// one bad file must not stop the others being updated
		if parseErr := p.parseFile(path, rel); parseErr != nil {
			log.Error().Err(parseErr).Msgf("failed to process %s", path)
			failures = append(failures, fmt.Errorf("%s: %w", path, parseErr))
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("directory walk failed: %w", err)
	}

	return errors.Join(failures...)
}

// parseFile parses a single Dockerfile and writes the labeled version. rel is the
// file's path below the scanned directory, used to lay out Output.
func (p *Parser) parseFile(filePath, rel string) error {
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

	outputPath := filePath
	mode := os.FileMode(0o644)

	if info, statErr := os.Stat(filePath); statErr == nil {
		mode = info.Mode().Perm()
	}

	if p.Output != "" {
		outputPath = filepath.Join(p.Output, rel)

		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil { //#nosec G301
			return fmt.Errorf("failed to create %s: %w", filepath.Dir(outputPath), err)
		}
	} else if string(dockerfile.Raw) == dump {
		log.Info().Msgf("unchanged: %s", filePath)

		return nil
	}

	//#nosec G306 G703 -- the path comes from the user's own arguments
	if err := os.WriteFile(outputPath, []byte(dump), mode); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	log.Info().Msgf("updated: %s", outputPath)

	return nil
}
