package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Github struct {
	Token     string
	RepoOwner string
}

func (g *Github) GetPublishedVersion(imageName string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/packages/container/%s/versions", g.RepoOwner, imageName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+g.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Received non-OK status code:", resp.StatusCode)
		return "", err
	}

	var data []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Error decoding response body:", err)
		return "", err
	}

	for _, image := range data {
		if metadata, ok := image["metadata"].(map[string]interface{}); ok {
			if container, ok := metadata["container"].(map[string]interface{}); ok {
				if tags, ok := container["tags"].([]interface{}); ok {
					for i, tag := range tags {
						if tagStr, ok := tag.(string); ok && tagStr == "latest" {
							tags = append(tags[:i], tags[i+1:]...)
							longest := ""
							for _, t := range tags {
								if tStr, ok := t.(string); ok && len(tStr) > len(longest) {
									longest = tStr
								}
							}
							return longest, nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("no data to go through")
}
