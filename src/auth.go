package stevedore

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func GetAuthToken(from string) (string, error) {
	url := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + from + ":pull"
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	res, _ := http.DefaultClient.Do(req)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("failed to close http client")
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body %w", err)
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(body, &jsonMap)

	if err != nil {
		return "", fmt.Errorf("marshal failure: %w", err)
	}

	token, ok := jsonMap["token"].(string)
	if !ok {
		return "", fmt.Errorf("failed to assert %s", jsonMap["token"])
	}

	return token, nil
}
