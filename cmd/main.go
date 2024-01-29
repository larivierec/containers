package main

import (
	"encoding/json"
	"fmt"
	"larivierec/containers/m/v2/pkg/api"
	"larivierec/containers/m/v2/pkg/provider"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Interface

var githubApi = &provider.Github{}

// Input Structs

type Channel struct {
	Name      string   `yaml:"name"`
	Platforms []string `yaml:"platforms"`
	Stable    bool     `yaml:"stable"`
}

type Metadata struct {
	App      string    `yaml:"app"`
	Url      string    `yaml:"url"`
	Rules    []string  `yaml:"rules"`
	Channels []Channel `yaml:"channels"`
}

// Output Structs
type Image struct {
	Name             string   `json:"name"`
	Version          string   `json:"version"`
	PublishedVersion string   `json:"published_version,omitempty"`
	Tags             []string `json:"tags"`
	LabelType        string   `json:"label_type"`
}

type Platform struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	Platform       string `json:"platform"`
	TargetOS       string `json:"target_os"`
	TargetArch     string `json:"target_arch"`
	Channel        string `json:"channel"`
	DockerfilePath string `json:"dockerfile"`
	DockerContext  string `json:"context"`
	LabelType      string `json:"label_type"`
}

type ImagesToBuild struct {
	ImagePlatforms []Platform `json:"image_platforms"`
	Images         []Image    `json:"images"`
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

func getLatestVersion(subDir, channelName string) string {
	ciDir := filepath.Join(subDir, "ci")
	if fileInfo, err := os.Stat(filepath.Join(ciDir, "latest.sh")); err == nil && !fileInfo.IsDir() {
		return getLatestVersionSh(filepath.Join(ciDir, "latest.sh"), channelName)
	} else if fileInfo, err := os.Stat(filepath.Join(subDir, channelName, "latest.sh")); err == nil && !fileInfo.IsDir() {
		return getLatestVersionSh(filepath.Join(subDir, channelName, "latest.sh"), channelName)
	}
	return ""
}

func getPlatformMetadata(subDir string, meta Metadata, forRelease bool, force bool, call api.Interface, channels []string) *ImagesToBuild {
	imagesToBuild := &ImagesToBuild{}
	filteredChannels := []Channel{}

	if len(channels) == 0 {
		filteredChannels = append(filteredChannels, meta.Channels...)
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
		// call.GetPublishedReleases(meta.Url, meta.Rules)
		version := getLatestVersion(subDir, channelName)
		if version == "" {
			continue
		}

		toBuild := &Image{}
		if channel.Stable {
			toBuild.Name = meta.App
		} else {
			toBuild.Name = fmt.Sprintf("%s-%s", meta.App, channel.Name)
		}

		if !force {
			published, resp, err := call.GetPublishedVersion(toBuild.Name)
			if ((err == nil || resp.StatusCode == http.StatusNotFound) && published != "") && strings.Contains(published, version) {
				continue
			}
			toBuild.PublishedVersion = published
		}
		toBuild.Version = version
		toBuild.Tags = []string{"latest", version}
		toBuild.LabelType = "org.opencontainers.image"

		for _, platform := range channel.Platforms {
			targetOs := strings.Split(platform, "/")[0]
			targetArch := strings.Split(platform, "/")[1]

			platformObj := Platform{
				Name:       toBuild.Name,
				Channel:    channel.Name,
				TargetOS:   targetOs,
				TargetArch: targetArch,
				Platform:   platform,
				Version:    version,
				LabelType:  "org.opencontainers.image",
			}

			if fileInfo, err := os.Stat(filepath.Join(subDir, channel.Name, "Dockerfile")); err == nil && !fileInfo.IsDir() {
				platformObj.DockerfilePath = filepath.Join(subDir, channel.Name, "Dockerfile")
				platformObj.DockerContext = filepath.Join(subDir, channel.Name)
			} else {
				platformObj.DockerfilePath = filepath.Join(subDir, "Dockerfile")
				platformObj.DockerContext = subDir
			}
			imagesToBuild.ImagePlatforms = append(imagesToBuild.ImagePlatforms, platformObj)
		}
		imagesToBuild.Images = append(imagesToBuild.Images, *toBuild)
	}

	return imagesToBuild
}

func main() {
	apiInit()
	args := os.Args[1:]
	if len(args) < 3 {
		fmt.Println("Usage: go run cmd/main.go <apps> <forRelease> <force> [<channels>]")
		os.Exit(1)
	}

	apps := args[0]
	forRelease, _ := strconv.ParseBool(args[1])
	force, _ := strconv.ParseBool(args[2])
	var channels []string

	if len(args) > 3 {
		channels = strings.Split(args[3], ",")
	}

	imagesToBuild := ImagesToBuild{
		ImagePlatforms: []Platform{},
		Images:         []Image{},
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
	processSpecificApps(selectedApps, forRelease, force, channels, &imagesToBuild)

	output, err := json.Marshal(imagesToBuild)
	if err != nil {
		fmt.Println("Error marshaling imagesToBuild:", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}

func processSpecificApps(selectedApps []string, forRelease bool, force bool, channels []string, imagesToBuild *ImagesToBuild) {
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

		imageToBuild := getPlatformMetadata(appDir, *meta, forRelease, force, githubApi, channels)
		imagesToBuild.Images = append(imagesToBuild.Images, imageToBuild.Images...)
		imagesToBuild.ImagePlatforms = append(imagesToBuild.ImagePlatforms, imageToBuild.ImagePlatforms...)
	}
}

func apiInit() {
	githubApi.RepoOwner = os.Getenv("GITHUB_REPOSITORY_OWNER")
	githubApi.Token = os.Getenv("GITHUB_TOKEN")

	if githubApi.RepoOwner == "" {
		githubApi.RepoOwner = os.Getenv("REPO_OWNER")
	}

	if githubApi.Token == "" {
		githubApi.Token = os.Getenv("TOKEN")
	}
}
