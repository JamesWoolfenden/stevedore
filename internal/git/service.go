package git

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rs/zerolog/log"
)

// Service defines the Git operations the labeller needs
type Service interface {
	GetCommitHash() (string, error)
	GetOrganization() string
	GetRepoName() string
	GetRelativePath(absPath string) (string, error)
}

// GitService implements Git operations for the stevedore tool
type GitService struct {
	gitRootDir   string
	repository   *git.Repository
	organization string
	repoName     string
}

// NewGitService creates a new GitService instance by walking up the directory tree
// to find the git repository root
func NewGitService(rootDir string) (Service, error) {
	var repository *git.Repository
	var err error

	rootDirIter, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Walk up the directory tree to find .git directory
	for {
		repository, err = git.PlainOpen(rootDirIter)
		if err == nil {
			break
		}

		newRootDir := filepath.Dir(rootDirIter)

		if rootDirIter == newRootDir {
			break
		}

		rootDirIter = newRootDir
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	gitService := &GitService{
		gitRootDir: rootDirIter,
		repository: repository,
	}

	if err := gitService.setOrgAndName(); err != nil {
		return nil, err
	}

	return gitService, nil
}

// setOrgAndName extracts organization and repository name from the origin remote.
// The remote is optional metadata: if there is none, or its URL isn't of the form
// host/org/repo, org and repo are left empty rather than failing, so the commit
// hash is still available.
func (g *GitService) setOrgAndName() error {
	remotes, err := g.repository.Remotes()
	if err != nil {
		return fmt.Errorf("failed to get remotes: %w", err)
	}

	for _, remote := range remotes {
		cfg := remote.Config()
		if cfg.Name != "origin" || len(cfg.URLs) == 0 {
			continue
		}

		// endpoint is structured like '/github.com/bridgecrewio/yor.git'
		endpoint, err := transport.NewEndpoint(cfg.URLs[0])
		if err != nil {
			log.Warn().Err(err).Msgf("cannot parse origin %q, skipping repo name", cfg.URLs[0])

			return nil
		}

		// a local path (file endpoint) has no host and so no org/repo to speak of
		if endpoint.Host == "" {
			log.Warn().Msgf("origin %q has no host, skipping repo name", cfg.URLs[0])

			return nil
		}

		// remove leading '/' from path and trailing '.git. suffix, then split by '/'
		parts := strings.Split(strings.TrimSuffix(strings.TrimLeft(endpoint.Path, "/"), ".git"), "/")
		if len(parts) < 2 || parts[0] == "" {
			log.Warn().Msgf("origin %q is not host/org/repo, skipping repo name", cfg.URLs[0])

			return nil
		}

		g.organization = parts[0]
		g.repoName = strings.Join(parts[1:], "/")

		break
	}

	return nil
}

// GetCommitHash returns the current HEAD commit hash
func (g *GitService) GetCommitHash() (string, error) {
	ref, err := g.repository.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	return ref.Hash().String(), nil
}

// GetOrganization returns the git organization name
func (g *GitService) GetOrganization() string {
	return g.organization
}

// GetRepoName returns the repository name
func (g *GitService) GetRepoName() string {
	return g.repoName
}

// GetRelativePath computes the relative file path from the git root
func (g *GitService) GetRelativePath(absPath string) (string, error) {
	relPath, err := filepath.Rel(g.gitRootDir, absPath)
	if err != nil {
		return "", fmt.Errorf("failed to compute relative path: %w", err)
	}

	return relPath, nil
}
