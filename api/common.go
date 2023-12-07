package api

type Interface interface {
	GetPublishedVersion(imageName string) (string, error)
}
