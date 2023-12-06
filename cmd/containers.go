package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Platform struct {
	Name           string `yaml:"name"`
	Version        string `yaml:"version"`
	Platform       string `yaml:"platform"`
	Channel        string `yaml:"channel"`
	DockerfilePath string `yaml:"dockerfile"`
	DockerContext  string `yaml:"context"`
	LabelType      string `yaml:"label_type"`
}

func (p *Platform) toMap() map[string]interface{} {
	return map[string]interface{}{
		"name":       p.Name,
		"version":    p.Version,
		"platform":   p.Platform,
		"channel":    p.Channel,
		"dockerfile": p.DockerfilePath,
		"context":    p.DockerContext,
		"label_type": p.LabelType,
	}
}

type Channel struct {
	Name      string   `yaml:"name"`
	Platforms []string `yaml:"platforms"`
	Stable    bool     `yaml:"stable"`
}

type Metadata struct {
	App      string    `yaml:"app"`
	Channels []Channel `yaml:"channels"`
}

func loadMetadataFileYAML(filePath string) (*Metadata, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var metadata Metadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func getLatestVersionSh(latestShPath string, channelName string) string {
	out, err := exec.Command(latestShPath, channelName).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getLatestVersion(subdir, channelName string) string {
	ciDir := filepath.Join(subdir, "ci")
	if fileInfo, err := os.Stat(filepath.Join(ciDir, "latest.sh")); err == nil && !fileInfo.IsDir() {
		return getLatestVersionSh(filepath.Join(ciDir, "latest.sh"), channelName)
	} else if fileInfo, err := os.Stat(filepath.Join(subdir, channelName, "latest.sh")); err == nil && !fileInfo.IsDir() {
		return getLatestVersionSh(filepath.Join(subdir, channelName, "latest.sh"), channelName)
	}
	return ""
}

func getPublishedVersion(imageName string) string {
	// Implement your getPublishedVersion logic here
	return "" // Replace with your actual logic
}

func getPlatformMetadata(subdir string, meta Metadata, forRelease, force bool, channels []string) map[string]interface{} {
	imagesToBuild := map[string]interface{}{
		"images":         []map[string]interface{}{},
		"imagePlatforms": []map[string]interface{}{},
	}

	filteredChannels := []Channel{}

	if len(channels) == 0 {
		for _, channel := range meta.Channels {
			filteredChannels = append(filteredChannels, channel)
		}
	} else {
		for _, channel := range meta.Channels {
			for _, channelName := range channels {
				if channel.Name == channelName {
					filteredChannels = append(filteredChannels, channel)
				}
			}
		}
	}

	for _, channel := range filteredChannels {
		channelName := channel.Name
		version := getLatestVersion(subdir, channelName)
		if version == "" {
			continue
		}

		toBuild := map[string]interface{}{}
		if channel.Stable {
			toBuild["name"] = meta.App
		} else {
			toBuild["name"] = fmt.Sprintf("%s-%s", meta.App, channel.Name)
		}

		// published := getPublishedVersion(toBuild["name"])
		// if !force && published != "" && published == version {
		// 	continue
		// }

		// toBuild["published_version"] = published
		toBuild["version"] = version
		toBuild["tags"] = []string{"rolling", version}
		toBuild["label_type"] = "org.opencontainers.image"

		for _, platform := range channel.Platforms {
			platformObj := Platform{}
			platformObj.Name = toBuild["name"].(string)
			platformObj.Channel = channel.Name
			platformObj.Platform = platform
			platformObj.Version = version
			platformObj.LabelType = "org.opencontainers.image"

			if fileInfo, err := os.Stat(filepath.Join(subdir, channel.Name, "Dockerfile")); err == nil && !fileInfo.IsDir() {
				platformObj.DockerfilePath = filepath.Join(subdir, channel.Name, "Dockerfile")
				platformObj.DockerContext = filepath.Join(subdir, channel.Name)
			} else {
				platformObj.DockerfilePath = filepath.Join(subdir, "Dockerfile")
				platformObj.DockerContext = subdir
			}
			imagesToBuild["imagePlatforms"] = append(imagesToBuild["imagePlatforms"].([]map[string]interface{}), platformObj.toMap())
		}
		imagesToBuild["images"] = append(imagesToBuild["images"].([]map[string]interface{}), toBuild)
	}

	return imagesToBuild
}

func main() {
	args := os.Args[1:]
	if len(args) < 3 {
		fmt.Println("Usage: go run cmd/containers.go <apps> <forRelease> <force> [<channels>]")
		os.Exit(1)
	}

	apps := args[0]
	forRelease, _ := strconv.ParseBool(args[1])
	force, _ := strconv.ParseBool(args[2])
	var channels []string

	if len(args) > 3 {
		channels = strings.Split(args[3], ",")
	}

	imagesToBuild := map[string]interface{}{
		"images":         []map[string]interface{}{},
		"imagePlatforms": []map[string]interface{}{},
	}

	selectedApps := []string{}
	if apps != "all" {
		selectedApps = strings.Split(apps, ",")
	} else {
		entries, err := os.ReadDir("./apps")
		if err != nil {
			log.Fatal(err)
		}
		for _, app := range entries {
			selectedApps = append(selectedApps, strings.Split(app.Name(), "/")[0])
		}
	}
	processSpecificApps(selectedApps, forRelease, force, channels, imagesToBuild)

	// Print or process the imagesToBuild map as needed
	output, err := json.Marshal(imagesToBuild)
	if err != nil {
		fmt.Println("Error marshaling imagesToBuild:", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}

func processSpecificApps(selectedApps []string, forRelease, force bool, channels []string, imagesToBuild map[string]interface{}) {
	for _, app := range selectedApps {
		appDir := "apps/" + app
		if _, err := os.Stat(appDir); os.IsNotExist(err) {
			fmt.Printf("app \"%s\" not found\n", app)
			continue
		}

		metaFile := appDir + "/ci/metadata.yaml"
		meta, err := loadMetadataFileYAML(metaFile)
		if err != nil {
			fmt.Printf("error loading metadata for app \"%s\": %v\n", app, err)
			continue
		}

		imageToBuild := getPlatformMetadata(appDir, *meta, forRelease, force, channels)
		if imageToBuild != nil {
			imagesToBuild["images"] = append(imagesToBuild["images"].([]map[string]interface{}), imageToBuild["images"].([]map[string]interface{})...)
			imagesToBuild["imagePlatforms"] = append(imagesToBuild["imagePlatforms"].([]map[string]interface{}), imageToBuild["imagePlatforms"].([]map[string]interface{})...)
		}
	}
}

func getMetadataForFile(subdir, file string) *Metadata {
	// Implement your logic to retrieve metadata from the file
	// You can determine the file type (yaml or json) and use the corresponding load function
	// Return the metadata as a Metadata struct pointer
	return nil // Replace with your actual logic
}
