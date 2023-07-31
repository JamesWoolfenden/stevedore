package stevedore

import (
	"fmt"
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

func Label(result *parser.Result, file *string) (string, error) {
	var label *parser.Node

	var endLine int

	var layer int64
	layer = 0

	myUser, _ := user.Current()

	if result == nil {
		return "", fmt.Errorf("dockerfile is nil")
	}

	for _, child := range result.AST.Children {
		endLine = child.EndLine

		if strings.Contains(child.Value, "FROM") {
			SplitFrom := strings.SplitN(child.Original, "FROM", 2)
			image := strings.TrimSpace(SplitFrom[1])
			ParentLabel, err := GetDockerLabels(image)

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
			label = MakeLabel(child, layer, myUser, endLine, file)
		}
	}

	if label == nil {
		var newLabel parser.Node

		MakeLabel(&newLabel, layer, myUser, endLine, file)

		result.AST.Children = append(result.AST.Children, &newLabel)
	}

	var dump string

	for _, child := range result.AST.Children {
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
			log.Error().Msgf("Failed to initialize git service for path \"%s\". Please ensure the provided root directory is initialized via the git init command: %q", absPath, err)
		}
	}

	url := strings.Split(gitService.remoteURL, "@")[1]
	url = strings.Replace(url, ":", "/", 1)
	url = strings.Replace(url, ".git", "", 1)

	url = "https://" + url

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
