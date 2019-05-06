package types

import "net/url"

// A Resource represents something that can be found on the web for the purposes of informing somone else in the community
type Resource struct {
	Type, Title string
	URL         url.URL
}

// A Repository is a place where resources are housed.
type Repository interface {
	Add(Resource) error
	Fetch(url.URL) (Resource, error)
}
