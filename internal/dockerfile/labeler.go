package dockerfile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jameswoolfenden/stevedore/internal/auth"
	"github.com/jameswoolfenden/stevedore/internal/config"
	"github.com/jameswoolfenden/stevedore/internal/git"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/rs/zerolog/log"
)

const dockerRegistryURL = "https://registry-1.docker.io/v2/"

// Dockerfile represents a parsed Dockerfile with metadata
type Dockerfile struct {
	Parsed *parser.Result
	Path   string
	Image  string
	Raw    []byte // the file as read, kept so labelling can preserve it
}

// Labeller handles adding labels to Dockerfiles
type Labeller struct {
	gitService  git.Service
	authService auth.DockerAuth
	httpClient  *http.Client
}

// NewLabeler creates a new Labeler instance
func NewLabeler(gitService git.Service, authService auth.DockerAuth) *Labeller {
	return &Labeller{
		gitService:  gitService,
		authService: authService,
		httpClient: &http.Client{
			Timeout: config.DefaultHTTPTimeout,
		},
	}
}

// ParseFile opens and parses a Dockerfile
func (d *Dockerfile) ParseFile() error {
	if err := config.ValidateDockerfilePath(d.Path); err != nil {
		return err
	}

	raw, err := os.ReadFile(d.Path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", d.Path, err)
	}

	log.Info().Msgf("opening: %s", d.Path)

	d.Raw = raw

	d.Parsed, err = parser.Parse(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("failed to parse dockerfile: %w", err)
	}

	return nil
}

var (
	// keys stevedore writes itself, and so may rewrite on a later run
	ownedKeyRe = regexp.MustCompile(`\s+(?:layer\.0\.(?:author|trace|tool)|git_(?:repo|org|file|commit))="(?:[^"\\]|\\.)*"`)
	traceRe    = regexp.MustCompile(`layer\.0\.trace="((?:[^"\\]|\\.)*)"`)
	quoter     = strings.NewReplacer(`\`, `\\`, `"`, `\"`)
)

// quote renders a LABEL value, escaping backslashes and double quotes
func quote(v string) string {
	return `"` + quoter.Replace(v) + `"`
}

// nodeText returns the source lines an AST node spans
func nodeText(lines []string, n *parser.Node) string {
	start, end := max(n.StartLine-1, 0), min(n.EndLine, len(lines))
	if start >= end {
		return ""
	}

	return strings.Join(lines[start:end], "\n")
}

// Label adds metadata labels to the Dockerfile.
//
// The original text is kept as it is: stevedore's keys are appended to an
// existing LABEL instruction (rewriting its own keys, and keeping the existing
// trace id, if it labelled the file before), or a new LABEL instruction is
// inserted after the last instruction.
func (l *Labeller) Label(dockerfile *Dockerfile, authorOverride string) (string, error) {
	if dockerfile == nil || dockerfile.Parsed == nil {
		return "", fmt.Errorf("dockerfile is nil")
	}

	raw := dockerfile.Raw
	if raw == nil {
		var err error

		raw, err = os.ReadFile(dockerfile.Path)
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", dockerfile.Path, err)
		}
	}

	eol := "\n"
	if bytes.Contains(raw, []byte("\r\n")) {
		eol = "\r\n"
	}

	lines := strings.Split(strings.TrimSuffix(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n"), "\n")

	// Get author information
	myUser, err := user.Current()
	if err != nil {
		log.Warn().Err(err).Msg("failed to get current user, using default")

		myUser = &user.User{Name: "unknown"}
	}

	if authorOverride != "" && authorOverride != "." {
		myUser.Name = authorOverride
	}

	var owned, lastLabel *parser.Node

	lastEnd := 0

	for _, child := range dockerfile.Parsed.AST.Children {
		lastEnd = max(lastEnd, child.EndLine)

		if strings.EqualFold(child.Value, "label") {
			lastLabel = child

			if traceRe.MatchString(nodeText(lines, child)) {
				owned = child
			}
		}
	}

	trace := ""

	if owned != nil {
		if m := traceRe.FindStringSubmatch(nodeText(lines, owned)); m != nil {
			trace = m[1]
		}
	}

	if trace == "" {
		trace = uuid.NewString()
	}

	keys := " layer.0.author=" + quote(myUser.Name) +
		" layer.0.trace=" + quote(trace) +
		` layer.0.tool="stevedore"` +
		l.gitKeys(dockerfile.Path)

	switch {
	case owned != nil:
		for i := max(owned.StartLine-1, 0); i < min(owned.EndLine, len(lines)); i++ {
			lines[i] = ownedKeyRe.ReplaceAllString(lines[i], "")
		}

		end := min(owned.EndLine, len(lines)) - 1
		lines[end] = strings.TrimRight(lines[end], " \t") + keys
	case lastLabel != nil:
		end := min(lastLabel.EndLine, len(lines)) - 1
		lines[end] = strings.TrimRight(lines[end], " \t") + keys
	default:
		lines = slices.Insert(lines, min(lastEnd, len(lines)), "LABEL"+keys)
	}

	log.Info().Msgf("file: %s", dockerfile.Path)

	return strings.Join(lines, eol) + eol, nil
}

// gitKeys returns the git metadata keys for a file, or "" if there is none
func (l *Labeller) gitKeys(filePath string) string {
	if l.gitService == nil {
		log.Debug().Msg("git service not available, skipping git metadata")

		return ""
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Warn().Err(err).Msgf("failed to get absolute path for %s", filePath)

		return ""
	}

	// Get commit hash from the existing repository (not cloning!)
	hash, err := l.gitService.GetCommitHash()
	if err != nil {
		log.Warn().Err(err).Msg("failed to get git commit hash")

		return ""
	}

	relPath, err := l.gitService.GetRelativePath(absPath)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get relative path")

		relPath = filepath.Base(absPath)
	}

	var keys strings.Builder

	if repo := l.gitService.GetRepoName(); repo != "" {
		keys.WriteString(" git_repo=" + quote(repo))
	}

	if org := l.gitService.GetOrganization(); org != "" {
		keys.WriteString(" git_org=" + quote(org))
	}

	keys.WriteString(" git_file=" + quote(filepath.ToSlash(relPath)))
	keys.WriteString(" git_commit=" + quote(hash))

	return keys.String()
}

// GetDockerLabels retrieves labels from a parent Docker image
func (l *Labeller) GetDockerLabels(dockerfile *Dockerfile) (map[string]interface{}, error) {
	image, version, err := splitImage(dockerfile.Image)
	if err != nil {
		return nil, err
	}

	// Get authentication token
	token, err := l.authService.GetAuthToken(image)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Fetch parent labels
	parentLabels, err := l.getParentLabels(image, version, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent labels for %s: %w", image, err)
	}

	return parentLabels, nil
}

// splitImage breaks an image reference into a Docker Hub repository and a tag or
// digest. Other registries are rejected, as the token flow here is Docker Hub's.
func splitImage(ref string) (repo, reference string, err error) {
	if ref == "" {
		return "", "", fmt.Errorf("image name cannot be empty")
	}

	reference = "latest"

	digested := false

	if before, digest, ok := strings.Cut(ref, "@"); ok {
		ref, reference, digested = before, digest, true
	}

	// a tag is a colon after the last slash; an earlier colon is a registry port
	if i := strings.LastIndex(ref, ":"); i > strings.LastIndex(ref, "/") {
		if !digested {
			reference = ref[i+1:]
		}

		ref = ref[:i]
	}

	first, rest, hasSlash := strings.Cut(ref, "/")
	if hasSlash && (first == "docker.io" || first == "index.docker.io") {
		ref, first, hasSlash = rest, "", strings.Contains(rest, "/")
		if hasSlash {
			first, _, _ = strings.Cut(rest, "/")
		}
	}

	if hasSlash && (strings.ContainsAny(first, ".:") || first == "localhost") {
		return "", "", fmt.Errorf("unsupported registry %q: only Docker Hub images are supported", first)
	}

	if !hasSlash {
		ref = "library/" + ref
	}

	return ref, reference, nil
}

// getParentLabels fetches labels from the Docker registry manifest
func (l *Labeller) getParentLabels(image, version, token string) (map[string]interface{}, error) {
	url := dockerRegistryURL + image + "/manifests/" + version

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set accept headers for different manifest formats
	req.Header.Add("accept", "application/vnd.oci.image.index.v1+json")
	req.Header.Add("accept", "application/vnd.oci.image.manifest.v1+json")
	req.Header.Add("accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := l.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}

	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close http response body")
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned %s for %s", res.Status, url)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var parentContainer map[string]interface{}
	if err := json.Unmarshal(body, &parentContainer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.Debug().Interface("parent_container", parentContainer).Msg("fetched parent container manifest")

	// Extract labels from history
	history, ok := parentContainer["history"].([]interface{})
	if !ok {
		log.Debug().Msg("no history entry in parent container")
		return map[string]interface{}{}, nil
	}

	if len(history) == 0 {
		return map[string]interface{}{}, nil
	}

	temp, ok := history[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("history entry is not map[string]interface{}")
	}

	previous, ok := temp["v1Compatibility"].(string)
	if !ok {
		log.Debug().Msg("no v1Compatibility in history")
		return map[string]interface{}{}, nil
	}

	var parent map[string]interface{}
	if err := json.Unmarshal([]byte(previous), &parent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal v1Compatibility: %w", err)
	}

	config, ok := parent["container_config"].(map[string]interface{})
	if !ok {
		log.Debug().Msg("no container_config in parent")
		return map[string]interface{}{}, nil
	}

	parentLabels, ok := config["Labels"].(map[string]interface{})
	if !ok {
		log.Debug().Msg("no labels in container_config")
		return map[string]interface{}{}, nil
	}

	return parentLabels, nil
}
