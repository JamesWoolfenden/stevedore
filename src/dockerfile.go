package stevedore

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

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/uuid"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/rs/zerolog/log"
)

type Dockerfile struct {
	Parsed *parser.Result
	Path   string
	Image  string
}

func (result *Dockerfile) ParseFile() error {
	data, err := os.Open(result.Path)
	log.Info().Msgf("opening: %s", result.Path)

	if err != nil {
		return fmt.Errorf("readfile error: %w", err)
	}

	result.Parsed, err = parser.Parse(data)

	defer func(data *os.File) {
		err := data.Close()
		if err != nil {
			log.Fatal().Msgf("close error:%s", err)
		}
	}(data)

	return err
}

func (result *Dockerfile) Label(Author string) (string, error) {
	var label *parser.Node

	var endLine int

	var layer int64
	layer = 0

	myUser, _ := user.Current()
	if Author != "" {
		myUser.Name = Author
	}

	if result.Parsed == nil {
		return "", fmt.Errorf("dockerfile is nil")
	}

	for _, child := range result.Parsed.AST.Children {
		endLine = child.EndLine

		if strings.Contains(child.Value, "FROM") {
			SplitFrom := strings.SplitN(child.Original, "FROM", 2)
			result.Image = strings.TrimSpace(SplitFrom[1])
			ParentLabel, err := result.GetDockerLabels()

			if err != nil {
				log.Info().Msgf("label error: %s", err)
			}

			for pLabel := range ParentLabel {
				if strings.Contains(pLabel, "layer") {
					splitter := strings.Split(pLabel, ".")
					version, _ := strconv.Atoi(splitter[1])
					layer = int64(version) + 1

					break
				}
			}
		}

		if strings.Contains(child.Value, "LABEL") {
			label = MakeLabel(child, layer, myUser, endLine, &result.Path)
		}
	}

	if label == nil {
		var newLabel parser.Node

		MakeLabel(&newLabel, layer, myUser, endLine, &result.Path)

		result.Parsed.AST.Children = append(result.Parsed.AST.Children, &newLabel)
	}

	var dump string

	for _, child := range result.Parsed.AST.Children {
		dump += child.Original + "\n"
	}

	return dump, nil
}

func MakeLabel(child *parser.Node, layer int64, myUser *user.User, endLine int, file *string) *parser.Node {
	var err error

	myLayer := " layer." + strconv.FormatInt(layer, 10)
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

	// dockerfile as absolute path
	absPath, err := filepath.Abs(*file)
	if err != nil {
		log.Info().Msgf("%s", err)
	}

	var gitService *GitService

	if absPath != "" {
		gitService, err = NewGitService(absPath)
		if err != nil {
			log.Error().Msgf("Failed to initialize git service for path \"%s\". Please ensure the provided root directory is initialized via the git init command: %q", absPath, err) //nolint:lll
		}
	}

	var url string

	if strings.Contains(gitService.remoteURL, "@") {
		url = strings.Split(gitService.remoteURL, "@")[1]
		url = strings.Replace(url, ":", "/", 1)
		url = strings.Replace(url, ".git", "", 1)

		url = "https://" + url
	} else {
		url = gitService.remoteURL
	}

	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: url,
	})

	//handle err

	// ... retrieves the branch pointed by HEAD
	ref, err := r.Head()

	//handle err

	hash := ref.Hash().String()

	if gitService != nil {
		child.Original += " git_repo=" + "\"" + gitService.repoName + "\""
		child.Original += " git_org=" + "\"" + gitService.organization + "\""
		child.Original += " git_file=" + "\"" + gitService.scanPathFromRoot + "\""
		child.Original += " git_commit=" + "\"" + hash + "\""
	}

	log.Info().Msgf("file: " + *file)
	log.Info().Msgf("label: " + child.Original)

	return child
}

func (result *Dockerfile) GetDockerLabels() (map[string]interface{}, error) {
	version := "latest"
	splitter := strings.SplitN(result.Image, "/", 2)

	if len(splitter) < 2 {
		result.Image = "library/" + result.Image
	}

	splitterVersion := strings.SplitN(result.Image, ":", 2)
	if len(splitterVersion) == 2 {
		version = splitterVersion[1]
		result.Image = splitterVersion[0]
	}

	token, err2 := GetAuthToken(result.Image)

	if err2 != nil {
		return nil, err2
	}

	ParentLabels, err := getParentLabels(result.Image, version, token)
	if err != nil {
		return nil, fmt.Errorf("get ParentLabels failed %w", err)
	}

	return ParentLabels, nil
}

func getParentLabels(from string, version string, token string) (map[string]interface{}, error) {
	url := "https://registry-1.docker.io/v2/" + from + "/manifests/" + version
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http query %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Info().Msgf("failed to close http client")
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	parentContainer := make(map[string]interface{})
	err = json.Unmarshal(body, &parentContainer)

	if err != nil {
		return nil, fmt.Errorf("marshalling fail %w", err)
	}

	history, ok := parentContainer["history"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("history entry no in parent container")
	}

	temp, ok := history[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("history entry not is not map[string]interface{}")
	}

	previous, ok := temp["v1Compatibility"].(string)
	if !ok {
		return nil, fmt.Errorf("v1Compatibility cannot be a string")
	}

	parent := make(map[string]interface{})
	err = json.Unmarshal([]byte(previous), &parent)

	if err != nil {
		return nil, fmt.Errorf("marshalling fail %w", err)
	}

	config, ok := parent["container_config"].(map[string]interface{})

	if !ok {
		log.Info().Msgf("no container config")
		return nil, nil
	}

	ParentLabels, ok := config["Labels"].(map[string]interface{})

	if !ok {
		log.Info().Msgf("no parent labels")
		return nil, nil
	}

	return ParentLabels, nil
}
