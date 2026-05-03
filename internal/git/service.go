package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rs/zerolog/log"
)

// Service defines the interface for Git operations
type Service interface {
	GetCommitHash() (string, error)
	GetOrganization() string
	GetRepoName() string
	GetRelativePath(absPath string) (string, error)
	GetRemoteURL() string
	GetFileBlame(filePath string) (*git.BlameResult, error)
	GetCurrentUserEmail() string
}

// GitService implements Git operations for the stevedore tool
type GitService struct {
	gitRootDir       string
	scanPathFromRoot string
	repository       *git.Repository
	remoteURL        string
	organization     string
	repoName         string
	blameByFile      *sync.Map
	currentUserEmail string
}

var gitGraphLock sync.Mutex

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

	scanAbsDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute scan path: %w", err)
	}

	scanPathFromRoot, err := filepath.Rel(rootDirIter, scanAbsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to compute relative path: %w", err)
	}

	gitService := &GitService{
		gitRootDir:       rootDirIter,
		scanPathFromRoot: scanPathFromRoot,
		repository:       repository,
		blameByFile:      &sync.Map{},
	}

	if err := gitService.setOrgAndName(); err != nil {
		return nil, err
	}

	gitService.currentUserEmail = getGitUserEmail()

	return gitService, nil
}

// setOrgAndName extracts organization and repository name from git remote URL
func (g *GitService) setOrgAndName() error {
	// get remotes to find the repository's url
	remotes, err := g.repository.Remotes()
	if err != nil {
		return fmt.Errorf("failed to get remotes: %w", err)
	}

	for _, remote := range remotes {
		if remote.Config().Name == "origin" {
			g.remoteURL = remote.Config().URLs[0]
			// get endpoint structured like '/github.com/bridgecrewio/yor.git
			endpoint, err := transport.NewEndpoint(g.remoteURL)
			if err != nil {
				return fmt.Errorf("failed to parse git endpoint: %w", err)
			}
			// remove leading '/' from path and trailing '.git. suffix, then split by '/'
			endpointPathParts := strings.Split(strings.TrimSuffix(strings.TrimLeft(endpoint.Path, "/"), ".git"), "/")
			if len(endpointPathParts) < 2 {
				return fmt.Errorf("invalid format of endpoint path: %s", endpoint.Path)
			}
			g.organization = endpointPathParts[0]
			g.repoName = strings.Join(endpointPathParts[1:], "/")

			break
		}
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

// GetRemoteURL returns the git remote URL
func (g *GitService) GetRemoteURL() string {
	return g.remoteURL
}

// GetRelativePath computes the relative file path from the git root
func (g *GitService) GetRelativePath(absPath string) (string, error) {
	relPath, err := filepath.Rel(g.gitRootDir, absPath)
	if err != nil {
		return "", fmt.Errorf("failed to compute relative path: %w", err)
	}
	return relPath, nil
}

// GetFileBlame retrieves git blame information for a file
func (g *GitService) GetFileBlame(filePath string) (*git.BlameResult, error) {
	blame, ok := g.blameByFile.Load(filePath)
	if ok {
		return blame.(*git.BlameResult), nil
	}

	relativeFilePath, err := g.GetRelativePath(filePath)
	if err != nil {
		return nil, err
	}

	var selectedCommit *object.Commit

	gitGraphLock.Lock() // Git is a graph, different files can lead to graph scans interfering with each other
	defer gitGraphLock.Unlock()

	head, err := g.repository.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository HEAD for file %s: %w", filePath, err)
	}

	selectedCommit, err = g.repository.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to find commit %s: %w", head.Hash().String(), err)
	}

	blameResult, err := git.Blame(selectedCommit, relativeFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get blame for latest commit of file %s: %w", filePath, err)
	}

	g.blameByFile.Store(filePath, blameResult)

	return blameResult, nil
}

// GetCurrentUserEmail returns the current git user's email
func (g *GitService) GetCurrentUserEmail() string {
	return g.currentUserEmail
}

// getGitUserEmail retrieves the git user email from git config
func getGitUserEmail() string {
	cmd := exec.Command("git", "config", "user.email")
	email, err := cmd.Output()

	if err != nil {
		log.Debug().Msgf("unable to get current git user email: %s", err)
		return ""
	}

	return strings.ReplaceAll(string(email), "\n", "")
}
