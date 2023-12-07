package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	tagHook string = "latest"
)

type Github struct {
	Token     string
	RepoOwner string
}

type Release struct {
	URL       string `json:"url,omitempty"`
	AssetsURL string `json:"assets_url,omitempty"`
	UploadURL string `json:"upload_url,omitempty"`
	HTMLURL   string `json:"html_url,omitempty"`
	ID        int    `json:"id,omitempty"`
	Author    struct {
		Login             string `json:"login,omitempty"`
		ID                int    `json:"id,omitempty"`
		NodeID            string `json:"node_id,omitempty"`
		AvatarURL         string `json:"avatar_url,omitempty"`
		GravatarID        string `json:"gravatar_id,omitempty"`
		URL               string `json:"url,omitempty"`
		HTMLURL           string `json:"html_url,omitempty"`
		FollowersURL      string `json:"followers_url,omitempty"`
		FollowingURL      string `json:"following_url,omitempty"`
		GistsURL          string `json:"gists_url,omitempty"`
		StarredURL        string `json:"starred_url,omitempty"`
		SubscriptionsURL  string `json:"subscriptions_url,omitempty"`
		OrganizationsURL  string `json:"organizations_url,omitempty"`
		ReposURL          string `json:"repos_url,omitempty"`
		EventsURL         string `json:"events_url,omitempty"`
		ReceivedEventsURL string `json:"received_events_url,omitempty"`
		Type              string `json:"type,omitempty"`
		SiteAdmin         bool   `json:"site_admin,omitempty"`
	} `json:"author,omitempty"`
	NodeID          string    `json:"node_id,omitempty"`
	TagName         string    `json:"tag_name,omitempty"`
	TargetCommitish string    `json:"target_commitish,omitempty"`
	Name            string    `json:"name,omitempty"`
	Draft           bool      `json:"draft,omitempty"`
	Prerelease      bool      `json:"prerelease,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	PublishedAt     time.Time `json:"published_at,omitempty"`
	Assets          []struct {
		URL      string `json:"url,omitempty"`
		ID       int    `json:"id,omitempty"`
		NodeID   string `json:"node_id,omitempty"`
		Name     string `json:"name,omitempty"`
		Label    string `json:"label,omitempty"`
		Uploader struct {
			Login             string `json:"login,omitempty"`
			ID                int    `json:"id,omitempty"`
			NodeID            string `json:"node_id,omitempty"`
			AvatarURL         string `json:"avatar_url,omitempty"`
			GravatarID        string `json:"gravatar_id,omitempty"`
			URL               string `json:"url,omitempty"`
			HTMLURL           string `json:"html_url,omitempty"`
			FollowersURL      string `json:"followers_url,omitempty"`
			FollowingURL      string `json:"following_url,omitempty"`
			GistsURL          string `json:"gists_url,omitempty"`
			StarredURL        string `json:"starred_url,omitempty"`
			SubscriptionsURL  string `json:"subscriptions_url,omitempty"`
			OrganizationsURL  string `json:"organizations_url,omitempty"`
			ReposURL          string `json:"repos_url,omitempty"`
			EventsURL         string `json:"events_url,omitempty"`
			ReceivedEventsURL string `json:"received_events_url,omitempty"`
			Type              string `json:"type,omitempty"`
			SiteAdmin         bool   `json:"site_admin,omitempty"`
		} `json:"uploader,omitempty"`
		ContentType        string    `json:"content_type,omitempty"`
		State              string    `json:"state,omitempty"`
		Size               int       `json:"size,omitempty"`
		DownloadCount      int       `json:"download_count,omitempty"`
		CreatedAt          time.Time `json:"created_at,omitempty"`
		UpdatedAt          time.Time `json:"updated_at,omitempty"`
		BrowserDownloadURL string    `json:"browser_download_url,omitempty"`
	} `json:"assets,omitempty"`
	TarballURL string `json:"tarball_url,omitempty"`
	ZipballURL string `json:"zipball_url,omitempty"`
	Body       string `json:"body,omitempty"`
	Reactions  struct {
		URL        string `json:"url,omitempty"`
		TotalCount int    `json:"total_count,omitempty"`
		Num1       int    `json:"+1,omitempty"`
		Num10      int    `json:"-1,omitempty"`
		Laugh      int    `json:"laugh,omitempty"`
		Hooray     int    `json:"hooray,omitempty"`
		Confused   int    `json:"confused,omitempty"`
		Heart      int    `json:"heart,omitempty"`
		Rocket     int    `json:"rocket,omitempty"`
		Eyes       int    `json:"eyes,omitempty"`
	} `json:"reactions,omitempty"`
}

type Package struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	PackageUrl string    `json:"package_html_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	HTMLUrl    string    `json:"html_url"`
	Metadata   struct {
		PackageType string `json:"package_type"`
		Container   struct {
			Tags []string `json:"tags"`
		} `json:"container"`
	} `json:"metadata"`
}

func (g *Github) GetPublishedVersion(imageName string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/packages/container/%s/versions", g.RepoOwner, imageName)
	resp, err := doRequest(url, g.Token)

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Received non-OK status code:", resp.StatusCode)
		return "", err
	}
	defer resp.Body.Close()
	images := []Package{}
	if err := json.NewDecoder(resp.Body).Decode(&images); err != nil {
		fmt.Println("Error decoding response body:", err)
		return "", err
	}

	for _, image := range images {
		for i, tag := range image.Metadata.Container.Tags {
			if tag == tagHook {
				image.Metadata.Container.Tags = append(image.Metadata.Container.Tags[:i], image.Metadata.Container.Tags[i+1:]...)
				longest := ""
				for _, t := range image.Metadata.Container.Tags {
					if len(t) > len(longest) {
						longest = t
					}
				}
				return longest, nil
			}
		}
	}

	return "", fmt.Errorf("no data to go through")
}

func (g *Github) GetPublishedReleases(url string, rules []string) ([]string, error) {
	resp, err := doRequest(url, g.Token)

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Received non-OK status code:", resp.StatusCode)
		return []string{}, err
	}
	defer resp.Body.Close()

	releases := []Release{}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		fmt.Println("Error decoding response body:", err)
		return []string{}, err
	}
	releases = filterReleases(releases, rules)
	for _, release := range releases {
		fmt.Println(release.Name)
	}
	return []string{}, err
}

func doRequest(url string, token string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	return resp, err
}

func filterReleases(releases []Release, rules []string) []Release {
	filteredReleases := []Release{}
	for _, release := range releases {
		for _, rule := range rules {
			if strings.Contains(strings.ToLower(release.Name), strings.ToLower(rule)) {
				filteredReleases = append(filteredReleases, release)
				continue
			}
		}
	}
	return filteredReleases
}
