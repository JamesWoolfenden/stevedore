package dockerfile

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// labelText writes content to a temp Dockerfile, labels it and returns the result.
func labelText(t *testing.T, content, author string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "Dockerfile")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	d := &Dockerfile{Path: path}
	if err := d.ParseFile(); err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	out, err := NewLabeler(nil, nil).Label(d, author)
	if err != nil {
		t.Fatalf("Label: %v", err)
	}

	return out
}

func isLabelLine(l string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(l)), "LABEL ")
}

func TestLabelPreservesOriginalText(t *testing.T) {
	t.Parallel()

	in := "# syntax=docker/dockerfile:1\n" +
		"# my comment\n" +
		"\n" +
		"FROM alpine:3.20\n" +
		"CMD [\"sh\"]\n" +
		"RUN <<EOF2\n" +
		"echo hi\n" +
		"EOF2\n"

	out := labelText(t, in, "tester")

	var kept []string

	labels := 0

	for _, l := range strings.Split(out, "\n") {
		if isLabelLine(l) {
			labels++

			if !strings.Contains(l, `layer.0.tool="stevedore"`) {
				t.Errorf("label line lacks tool key: %q", l)
			}

			continue
		}

		kept = append(kept, l)
	}

	if labels != 1 {
		t.Fatalf("want exactly 1 LABEL line, got %d in:\n%s", labels, out)
	}

	if got := strings.Join(kept, "\n"); got != in {
		t.Errorf("original text was altered.\nwant:\n%q\ngot:\n%q", in, got)
	}

	// the label must come after the heredoc terminator, not inside the body
	if strings.Index(out, "LABEL ") < strings.Index(out, "EOF2\n") {
		t.Errorf("LABEL inserted inside the heredoc:\n%s", out)
	}
}

func TestLabelKeepsCRLF(t *testing.T) {
	t.Parallel()

	out := labelText(t, "FROM alpine:3.20\r\nCMD [\"sh\"]\r\n", "tester")

	if strings.Contains(strings.ReplaceAll(out, "\r\n", ""), "\n") {
		t.Errorf("bare LF found in a CRLF file: %q", out)
	}
}

func TestLabelIsIdempotentAndKeepsTrace(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "Dockerfile")

	if err := os.WriteFile(path, []byte("FROM alpine:3.20\nCMD [\"sh\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	run := func(author string) string {
		d := &Dockerfile{Path: path}
		if err := d.ParseFile(); err != nil {
			t.Fatal(err)
		}

		out, err := NewLabeler(nil, nil).Label(d, author)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			t.Fatal(err)
		}

		return out
	}

	first := run("tester")
	second := run("tester")

	if first != second {
		t.Errorf("re-running changed the file.\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	traceRe := regexp.MustCompile(`layer\.0\.trace="([^"]+)"`)
	trace := traceRe.FindStringSubmatch(first)

	if trace == nil {
		t.Fatalf("no trace in %s", first)
	}

	third := run("someone-else")

	if got := traceRe.FindStringSubmatch(third); got == nil || got[1] != trace[1] {
		t.Errorf("trace uuid was not kept: %v vs %v", got, trace)
	}

	if !strings.Contains(third, `layer.0.author="someone-else"`) || strings.Contains(third, `layer.0.author="tester"`) {
		t.Errorf("author not updated in place:\n%s", third)
	}

	count := 0

	for _, l := range strings.Split(third, "\n") {
		if isLabelLine(l) {
			count++
		}
	}

	if count != 1 {
		t.Errorf("want 1 LABEL line after 3 runs, got %d:\n%s", count, third)
	}
}

func TestLabelAppendsToExistingLabel(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"lowercase":    "FROM alpine:3.20\nlabel a=b\nCMD [\"sh\"]\n",
		"continuation": "FROM alpine:3.20\nLABEL a=b \\\n      c=d\nCMD [\"sh\"]\n",
	}

	for name, in := range cases {
		in := in

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			out := labelText(t, in, "tester")

			count := 0

			for _, l := range strings.Split(out, "\n") {
				if isLabelLine(l) {
					count++
				}
			}

			if count != 1 {
				t.Errorf("want the existing LABEL extended, got %d LABEL lines:\n%s", count, out)
			}

			if !strings.Contains(out, `layer.0.author="tester"`) || !strings.Contains(out, "a=b") {
				t.Errorf("label content wrong:\n%s", out)
			}
		})
	}
}

func TestLabelEscapesValues(t *testing.T) {
	t.Parallel()

	out := labelText(t, "FROM alpine:3.20\n", `DOMAIN\us"er`)

	if !strings.Contains(out, `layer.0.author="DOMAIN\\us\"er"`) {
		t.Errorf("author not escaped:\n%s", out)
	}
}
