package storage

import (
	"context"
	"graphql_project/internal/graph/model"
)

type Storage interface {
	CreatePost(ctx context.Context, newPost model.NewPost) (*model.Post, error)
	GetAllPosts(ctx context.Context, offset *int, limit *int) ([]*model.Post, error)
	GetPostByID(ctx context.Context, id string) (*model.Post, error)
	CreateComment(ctx context.Context, newComment model.NewComment) (*model.Comment, error)
}
