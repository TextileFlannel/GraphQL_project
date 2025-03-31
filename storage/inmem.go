package storage

import (
	"context"
	"errors"
	"sync"
	"graphql_project/graph/model"
)

type InMemStorage struct {
	posts     map[string]*model.Post
	postsIDs  []string
	mu        sync.RWMutex
	postIDSeq int
}

func NewInMemStorage() *InMemStorage {
	return &InMemStorage{
		posts:    make(map[string]*model.Post),
		postsIDs: make([]string, 0),
	}
}

func (s *InMemStorage) CreatePost(ctx context.Context, post *model.Post) (*model.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.postIDSeq++
	post.ID = string(rune(s.postIDSeq))
	post.CommentsEnabled = true

	s.posts[post.ID] = post
	s.postsIDs = append(s.postsIDs, post.ID)

	return post, nil
}

func (s *InMemStorage) GetAllPosts(ctx context.Context, first int32, after *string) ([]*model.Post, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := 0
	if after != nil {
		for i, id := range s.postsIDs {
			if id == *after {
				start = i + 1
				break
			}
		}
	}

	end := start + int(first)
	if end > len(s.postsIDs) {
		end = len(s.postsIDs)
	}

	hasNext := end < len(s.postsIDs)
	ids := s.postsIDs[start:end]
	
	posts := make([]*model.Post, 0, len(ids))
	for _, id := range ids {
		posts = append(posts, s.posts[id])
	}

	return posts, hasNext, nil
}

func (s *InMemStorage) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	post, exists := s.posts[id]
	if !exists {
		return nil, ErrNotFound
	}
	return post, nil
}

func (s *InMemStorage) ToggleComments(ctx context.Context, postID string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, exists := s.posts[postID]
	if !exists {
		return ErrNotFound
	}

	post.CommentsEnabled = enabled
	return nil
}

var (
	ErrNotFound = errors.New("not found")
	ErrNotImplemented = errors.New("not implemented")
)