package api

import "net/http"

type Interface interface {
	GetPublishedVersion(imageName string) (string, *http.Response, error)
	GetPublishedReleases(url string, rules []string) ([]string, *http.Response, error)
}
