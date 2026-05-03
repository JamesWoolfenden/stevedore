package dockerfile

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
			Timeout: 30 * time.Second,
		},
	}
}

// ParseFile opens and parses a Dockerfile
func (d *Dockerfile) ParseFile() error {
	if err := config.ValidateDockerfilePath(d.Path); err != nil {
		return err
	}

	data, err := os.Open(d.Path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", d.Path, err)
	}

	defer func() {
		if closeErr := data.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msgf("failed to close file: %s", d.Path)
		}
	}()

	log.Info().Msgf("opening: %s", d.Path)

	d.Parsed, err = parser.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse dockerfile: %w", err)
	}

	return nil
}

// Label adds metadata labels to the Dockerfile
func (l *Labeller) Label(dockerfile *Dockerfile, authorOverride string) (string, error) {
	if dockerfile.Parsed == nil {
		return "", fmt.Errorf("dockerfile is nil")
	}

	var label *parser.Node
	var endLine int
	var layer int64

	// Get author information
	myUser, err := user.Current()
	if err != nil {
		log.Warn().Err(err).Msg("failed to get current user, using default")
		myUser = &user.User{Name: "unknown"}
	}

	if authorOverride != "" && authorOverride != "." {
		myUser.Name = authorOverride
	}

	// Process AST nodes to find or create LABEL instruction
	for _, child := range dockerfile.Parsed.AST.Children {
		endLine = child.EndLine

		if strings.Contains(child.Value, "LABEL") {
			label = l.makeLabel(child, layer, myUser, endLine, dockerfile.Path)
		}
	}

	// Create new label if none exists
	if label == nil {
		var newLabel parser.Node
		l.makeLabel(&newLabel, layer, myUser, endLine, dockerfile.Path)
		dockerfile.Parsed.AST.Children = append(dockerfile.Parsed.AST.Children, &newLabel)
	}

	// Build output from AST
	var dump strings.Builder
	for _, child := range dockerfile.Parsed.AST.Children {
		dump.WriteString(child.Original)
		dump.WriteString("\n")
	}

	return dump.String(), nil
}

// makeLabel creates a label node with metadata
func (l *Labeller) makeLabel(child *parser.Node, layer int64, myUser *user.User, endLine int, filePath string) *parser.Node {
	myLayer := " layer." + strconv.FormatInt(layer, 10)

	// Build base label
	if strings.Contains(child.Value, "LABEL") {
		child.Original = child.Original + myLayer +
			".author=" + "\"" + myUser.Name + "\"" + myLayer + ".trace=\"" + uuid.NewString() + "\""
	} else {
		child.Original = "LABEL" + myLayer +
			".author=" + "\"" + myUser.Name + "\"" + myLayer + ".trace=\"" + uuid.NewString() + "\""
	}

	child.Original += myLayer + ".tool=\"stevedore\""
	child.StartLine = endLine + 1
	child.EndLine = endLine + 1

	// Add git metadata if available
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Warn().Err(err).Msgf("failed to get absolute path for %s", filePath)
		return child
	}

	if l.gitService != nil {
		l.addGitMetadata(child, absPath)
	} else {
		log.Debug().Msg("git service not available, skipping git metadata")
	}

	log.Info().Msgf("file: %s", filePath)
	log.Info().Msgf("label: %s", child.Original)

	return child
}

// addGitMetadata adds git-related metadata to the label
func (l *Labeller) addGitMetadata(child *parser.Node, absPath string) {
	// Get commit hash from the existing repository (not cloning!)
	hash, err := l.gitService.GetCommitHash()
	if err != nil {
		log.Warn().Err(err).Msg("failed to get git commit hash")
		return
	}

	relPath, err := l.gitService.GetRelativePath(absPath)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get relative path")
		relPath = filepath.Base(absPath)
	}

	child.Original += " git_repo=" + "\"" + l.gitService.GetRepoName() + "\""
	child.Original += " git_org=" + "\"" + l.gitService.GetOrganization() + "\""
	child.Original += " git_file=" + "\"" + relPath + "\""
	child.Original += " git_commit=" + "\"" + hash + "\""
}

// GetDockerLabels retrieves labels from a parent Docker image
func (l *Labeller) GetDockerLabels(dockerfile *Dockerfile) (map[string]interface{}, error) {
	version := "latest"
	image := dockerfile.Image

	// Normalize image name
	splitter := strings.SplitN(image, "/", 2)
	if len(splitter) < 2 {
		image = "library/" + image
	}

	// Extract version if specified
	splitterVersion := strings.SplitN(image, ":", 2)
	if len(splitterVersion) == 2 {
		version = splitterVersion[1]
		image = splitterVersion[0]
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
		return nil, nil
	}

	if len(history) == 0 {
		return nil, nil
	}

	temp, ok := history[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("history entry is not map[string]interface{}")
	}

	previous, ok := temp["v1Compatibility"].(string)
	if !ok {
		log.Debug().Msg("no v1Compatibility in history")
		return nil, nil
	}

	var parent map[string]interface{}
	if err := json.Unmarshal([]byte(previous), &parent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal v1Compatibility: %w", err)
	}

	config, ok := parent["container_config"].(map[string]interface{})
	if !ok {
		log.Debug().Msg("no container_config in parent")
		return nil, nil
	}

	parentLabels, ok := config["Labels"].(map[string]interface{})
	if !ok {
		log.Debug().Msg("no labels in container_config")
		return nil, nil
	}

	return parentLabels, nil
}
