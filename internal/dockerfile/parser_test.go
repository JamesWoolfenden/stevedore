package dockerfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const plain = "FROM alpine:3.20\nCMD [\"sh\"]\n"

func write(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func labelled(s string) bool { return strings.Contains(s, `layer.0.tool="stevedore"`) }

func TestParseAllUpdatesInPlaceByDefault(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "Dockerfile"), plain)
	write(t, filepath.Join(root, "sub", "Dockerfile"), plain)

	p := NewParser(NewLabeler(nil, nil))
	p.Directory = root

	if err := p.ParseAll(); err != nil {
		t.Fatal(err)
	}

	for _, f := range []string{"Dockerfile", filepath.Join("sub", "Dockerfile")} {
		if !labelled(read(t, filepath.Join(root, f))) {
			t.Errorf("%s was not updated in place", f)
		}
	}

	if _, err := os.Stat("Dockerfile"); err == nil {
		t.Error("a Dockerfile was written to the working directory")
	}
}

func TestParseAllOutputMirrorsTree(t *testing.T) {
	root := t.TempDir()
	out := t.TempDir()
	write(t, filepath.Join(root, "a", "Dockerfile"), plain)
	write(t, filepath.Join(root, "b", "Dockerfile"), plain)

	p := NewParser(NewLabeler(nil, nil))
	p.Directory = root
	p.Output = out

	if err := p.ParseAll(); err != nil {
		t.Fatal(err)
	}

	for _, d := range []string{"a", "b"} {
		if !labelled(read(t, filepath.Join(out, d, "Dockerfile"))) {
			t.Errorf("%s/Dockerfile missing or unlabelled in output dir", d)
		}

		if read(t, filepath.Join(root, d, "Dockerfile")) != plain {
			t.Errorf("%s/Dockerfile source was modified despite --output", d)
		}
	}
}

func TestParseAllOnlyTouchesDockerfiles(t *testing.T) {
	root := t.TempDir()

	touched := []string{"Dockerfile", "Dockerfile.dev", "app.Dockerfile"}
	untouched := []string{
		"Dockerfile.md", "notes-Dockerfile.txt", "Dockerfile.bak",
		filepath.Join(".git", "Dockerfile"),
		filepath.Join("node_modules", "x", "Dockerfile"),
		filepath.Join("vendor", "x", "Dockerfile"),
	}

	for _, f := range append(append([]string{}, touched...), untouched...) {
		write(t, filepath.Join(root, f), plain)
	}

	p := NewParser(NewLabeler(nil, nil))
	p.Directory = root

	if err := p.ParseAll(); err != nil {
		t.Fatal(err)
	}

	for _, f := range touched {
		if !labelled(read(t, filepath.Join(root, f))) {
			t.Errorf("%s should have been labelled", f)
		}
	}

	for _, f := range untouched {
		if read(t, filepath.Join(root, f)) != plain {
			t.Errorf("%s should not have been touched", f)
		}
	}
}

func TestParseAllContinuesPastBadFile(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "a", "Dockerfile"), "")
	write(t, filepath.Join(root, "b", "Dockerfile"), plain)

	p := NewParser(NewLabeler(nil, nil))
	p.Directory = root

	if err := p.ParseAll(); err == nil {
		t.Error("expected an error for the unparseable file")
	}

	if !labelled(read(t, filepath.Join(root, "b", "Dockerfile"))) {
		t.Error("a bad file stopped later Dockerfiles being labelled")
	}
}
