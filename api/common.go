package api

type Interface interface {
	GetPublishedVersion(imageName string) (string, error)
	GetPublishedReleases(url string, rules []string) ([]string, error)
}
