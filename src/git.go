package stevedore

import (
	gitinternal "github.com/jameswoolfenden/stevedore/internal/git"
)

// GitService is a backward compatibility type alias
// Deprecated: Use internal/git.Service interface instead
type GitService = gitinternal.Service

// NewGitService creates a new git service - backward compatibility wrapper
// Deprecated: Use internal/git.NewGitService instead
func NewGitService(rootDir string) (gitinternal.Service, error) {
	return gitinternal.NewGitService(rootDir)
}
