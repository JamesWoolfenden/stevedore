package stevedore

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/rs/zerolog/log"
)

func GetDockerLabels(from string) (map[string]interface{}, error) {
	version := "latest"
	splitter := strings.SplitN(from, "/", 2)

	if len(splitter) < 2 {
		from = "library/" + from
	}

	splitterVersion := strings.SplitN(from, ":", 2)
	if len(splitterVersion) == 2 {
		version = splitterVersion[1]
		from = splitterVersion[0]
	}

	token, err2 := GetAuthToken(from)

	if err2 != nil {
		return nil, err2
	}

	ParentLabels, err := GetParentLabels(from, version, token)

	if err != nil {
		return nil, fmt.Errorf("get ParentLabels failed %w", err)
	}

	return ParentLabels, nil
}

func GetParentLabels(from string, version string, token string) (map[string]interface{}, error) {
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

	history := parentContainer["history"].([]interface{})
	temp := history[0].(map[string]interface{})
	previous := temp["v1Compatibility"].(string)
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

func ParseFile(file string) (*parser.Result, error) {
	data, err := os.Open(file)
	log.Info().Msgf("opening: %s", file)

	if err != nil {
		return nil, fmt.Errorf("readfile error: %w", err)
	}

	result, err := parser.Parse(data)

	defer func(data *os.File) {
		err := data.Close()
		if err != nil {
			log.Fatal().Msgf("close error:%s", err)
		}
	}(data)

	return result, err
}

func ParseAll(file *string, directory string) error {
	if file != nil {
		err2 := Parse(file)
		if err2 != nil {
			return err2
		}
	} else {
		err := filepath.Walk(directory,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() && strings.Contains(info.Name(), "Dockerfile") {
					err2 := Parse(&path)
					if err2 != nil {
						return err2
					}
				}

				return nil
			})
		if err != nil {
			return fmt.Errorf("walk path failed %w", err)
		}
	}

	return nil
}

func Parse(file *string) error {
	result, err := ParseFile(*file)

	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	dump := Label(result)

	err = os.WriteFile("Dockerfile", []byte(dump), 0644)
	if err != nil {
		return fmt.Errorf("writefile error: %w", err)
	}

	log.Info().Msgf("updated: %s", "Dockerfile")
	return nil
}
