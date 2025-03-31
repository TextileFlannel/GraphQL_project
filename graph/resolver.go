package graph

import "graphql_project/storage"

type Resolver struct{
	Storage storage.Storage
}

func NewResolver(storage storage.Storage) *Resolver {
	return &Resolver{
		Storage: storage,
	}
}