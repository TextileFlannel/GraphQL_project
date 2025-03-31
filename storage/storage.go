package storage

import (
	"context"
	"graphql_project/graph/model"
)

type Storage interface {
	CreatePost(ctx context.Context, post *model.Post) (*model.Post, error)
	GetAllPosts(ctx context.Context, first int32, after *string) ([]*model.Post, bool, error)
	GetPostByID(ctx context.Context, id string) (*model.Post, error)
	ToggleComments(ctx context.Context, postID string, enabled bool) error
}