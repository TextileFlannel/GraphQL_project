package graph

import (
	"graphql_project/graph/model"
	"graphql_project/storage"
	"sync"
)

type observer struct {
    ch  chan *model.Comment
    postID string
}

type Resolver struct{
	Storage storage.Storage
	observers []observer
	mu sync.Mutex
}

func NewResolver(storage storage.Storage) *Resolver {
	return &Resolver{
		Storage: storage,
		observers: make([]observer, 0),
	}
}
