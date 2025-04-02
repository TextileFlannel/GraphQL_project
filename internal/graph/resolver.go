package graph

import (
	"graphql_project/internal/graph/model"
	"graphql_project/internal/service"
	"sync"
)

type observer struct {
	ch     chan *model.Comment
	postID string
}

type Resolver struct {
	Service   *service.Service
	observers []observer
	mu        sync.Mutex
}

func NewResolver(serv *service.Service) *Resolver {
	return &Resolver{
		Service:   serv,
		observers: make([]observer, 0),
	}
}
