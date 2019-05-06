package repository

import (
	"net/url"

	"github.com/willmadison/resourceful/resourceful"
)

// InMemoryRepository stores its resources in memory.
type InMemoryRepository struct {
	store map[url.URL]resourceful.Resource
}

// Add places a new resource in the repository.
func (i *InMemoryRepository) Add(r resourceful.Resource) error {
	i.store[r.URL] = r
	return nil
}

// Fetch retrieves an existing resource from the repository.
func (i *InMemoryRepository) Fetch(u url.URL) (resourceful.Resource, error) {
	return i.store[u], nil
}

// NewInMemory constructs a new in memory repository
func NewInMemory() *InMemoryRepository {
	return &InMemoryRepository{make(map[url.URL]resourceful.Resource)}
}
