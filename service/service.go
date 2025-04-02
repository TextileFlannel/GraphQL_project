package service

import (
	"context"
	"graphql_project/graph/model"
	"graphql_project/storage"
)

type Service struct {
	storage storage.Storage
}

func NewService(storage storage.Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) CreatePost(ctx context.Context, newPost model.NewPost) (*model.Post, error) {
	model, err := s.storage.CreatePost(ctx, newPost)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (s *Service) GetAllPosts(ctx context.Context, offset *int, limit *int) ([]*model.Post, error) {
	model, err := s.storage.GetAllPosts(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (s *Service) GetPost(ctx context.Context, id string) (*model.Post, error) {
	model, err := s.storage.GetPostByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return model, nil
}

func (s *Service) CreateComment(ctx context.Context, newComment model.NewComment) (*model.Comment, error) {
	model, err := s.storage.CreateComment(ctx, newComment)
	if err != nil {
		return nil, err
	}
	return model, nil
}
