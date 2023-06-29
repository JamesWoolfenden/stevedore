package main

import (
	"encoding/json"
	"fmt"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	data, err := os.Open("./examples/with/Dockerfile")

	if err != nil {
		log.Fatalf("readfile error: ", err)
	}

	result, err := parser.Parse(data)

	defer func(data *os.File) {
		err := data.Close()
		if err != nil {
			log.Fatalf("close error: ", err)
		}
	}(data)

	var label *parser.Node
	var endLine int

	var layer int64
	layer = 0

	for _, child := range result.AST.Children {
		endLine = child.EndLine
		if strings.Contains(child.Value, "FROM") {
			SplitFrom := strings.SplitN(child.Original, "FROM", 2)
			image := strings.TrimSpace(SplitFrom[1])
			ParentLabel, err := GetDockerLabels(image)

			if err != nil {
				log.Fatalf("label error: ", err)
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
			child.Original += " layer." + strconv.FormatInt(layer, 10) + ".author = \"JamesWoolfenden\""
			label = child
			continue
		}
	}

	if label == nil {
		var newLabel parser.Node
		newLabel.Value = "LABEL"
		newLabel.Original = "LABEL layer." + strconv.FormatInt(layer, 10) + ".author = \"JamesWoolfenden\""
		newLabel.StartLine = endLine + 1
		newLabel.EndLine = endLine + 1
		var child parser.Node
		result.AST.AddChild(&child, newLabel.StartLine, newLabel.StartLine)
		result.AST.Children = append(result.AST.Children, &newLabel)
	}

	var dump string

	for _, child := range result.AST.Children {
		dump += child.Original + "\n"
	}

	err = os.WriteFile("Dockerfile", []byte(dump), 0644)
	if err != nil {
		log.Fatalf("Writefile error: ", err)
	}
}

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

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	parentContainer := make(map[string]interface{})
	err := json.Unmarshal(body, &parentContainer)

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

	config := parent["container_config"].(map[string]interface{})

	ParentLabels, ok := config["Labels"].(map[string]interface{})

	if ok {
		return ParentLabels, nil
	}

	return nil, nil
}

func GetAuthToken(from string) (string, error) {
	url := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + from + ":pull"
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	jsonMap := make(map[string]interface{})
	err := json.Unmarshal(body, &jsonMap)

	if err != nil {
		return "", fmt.Errorf("marshal failure: %w", err)
	}

	token := jsonMap["token"].(string)
	return token, nil
}
