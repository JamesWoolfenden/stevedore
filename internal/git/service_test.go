package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// newRepo makes a repo with one commit; remote, if set, becomes "origin".
func newRepo(t *testing.T, remote string) (dir, hash string) {
	t.Helper()

	dir = t.TempDir()

	repo, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}

	if remote != "" {
		if _, err := repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{remote}}); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "sub", "Dockerfile"), []byte("FROM alpine\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := wt.Add("sub/Dockerfile"); err != nil {
		t.Fatal(err)
	}

	h, err := wt.Commit("init", &gogit.CommitOptions{
		Author: &object.Signature{Name: "t", Email: "t@example.com", When: time.Now()},
	})
	if err != nil {
		t.Fatal(err)
	}

	return dir, h.String()
}

func TestNewGitServiceParsesRemotes(t *testing.T) {
	t.Parallel()

	tests := map[string][2]string{
		"https://github.com/acme/widgets.git":   {"acme", "widgets"},
		"https://github.com/acme/widgets":       {"acme", "widgets"},
		"git@github.com:acme/widgets.git":       {"acme", "widgets"},
		"https://gitlab.com/group/sub/proj.git": {"group", "sub/proj"},
	}

	for remote, want := range tests {
		dir, hash := newRepo(t, remote)

		svc, err := NewGitService(dir)
		if err != nil {
			t.Fatalf("%s: %v", remote, err)
		}

		if svc.GetOrganization() != want[0] || svc.GetRepoName() != want[1] {
			t.Errorf("%s: got org %q repo %q, want %q %q", remote, svc.GetOrganization(), svc.GetRepoName(), want[0], want[1])
		}

		if got, err := svc.GetCommitHash(); err != nil || got != hash {
			t.Errorf("%s: hash = %q, %v; want %q", remote, got, err, hash)
		}
	}
}

// A repo with no usable origin must still give a commit hash; only org/repo are blank.
func TestNewGitServiceToleratesMissingOrOddOrigin(t *testing.T) {
	t.Parallel()

	for name, remote := range map[string]string{"none": "", "local path": "/srv/repo", "relative": "../x"} {
		dir, hash := newRepo(t, remote)

		svc, err := NewGitService(dir)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", name, err)
		}

		if svc.GetOrganization() != "" || svc.GetRepoName() != "" {
			t.Errorf("%s: want empty org/repo, got %q/%q", name, svc.GetOrganization(), svc.GetRepoName())
		}

		if got, err := svc.GetCommitHash(); err != nil || got != hash {
			t.Errorf("%s: hash = %q, %v; want %q", name, got, err, hash)
		}
	}
}

func TestNewGitServiceFromSubdirectory(t *testing.T) {
	t.Parallel()

	dir, _ := newRepo(t, "https://github.com/acme/widgets.git")

	svc, err := NewGitService(filepath.Join(dir, "sub"))
	if err != nil {
		t.Fatal(err)
	}

	rel, err := svc.GetRelativePath(filepath.Join(dir, "sub", "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}

	if filepath.ToSlash(rel) != "sub/Dockerfile" {
		t.Errorf("relative path = %q, want sub/Dockerfile", rel)
	}
}

func TestNewGitServiceOutsideRepo(t *testing.T) {
	t.Parallel()

	if _, err := NewGitService(t.TempDir()); err == nil {
		t.Skip("temp dir is inside a git repository on this machine")
	}
}
