package storage

import (
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/google/uuid"
	"graphql_project/internal/graph/model"
)

var (
	ErrNotCommentable = errors.New("the post is not commentable")
	ErrNotFound       = errors.New("not found")
	ErrBadRequest     = errors.New("bad request")
)

type inmemStorage struct {
	posts []*model.Post
	mu sync.RWMutex
}

func NewInMemStorage() *inmemStorage {
	return &inmemStorage{posts: make([]*model.Post, 0)}
}

func (s *inmemStorage) CreatePost(ctx context.Context, newPost model.NewPost) (*model.Post, error) {
	post := &model.Post{
		ID:          uuid.New(),
		Title:       newPost.Title,
		Author:      newPost.Author,
		Content:     newPost.Content,
		Commentable: newPost.Commentable,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.posts = append(s.posts, post)
	return post, nil
}

func (s *inmemStorage) GetAllPosts(ctx context.Context, offset *int, limit *int) ([]*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var off int
	if offset != nil && *offset < len(s.posts) {
		off = *offset
	}
	lim := len(s.posts)
	if limit != nil && *limit <= len(s.posts) {
		lim = *limit
	}

	if len(s.posts) == 0 {
		return s.posts, nil
	}
	return s.posts[off:lim], nil
}

func (s *inmemStorage) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	idx := slices.IndexFunc(s.posts, func(post *model.Post) bool {
		return post.ID.String() == id
	})
	if idx == -1 {
		return nil, ErrNotFound
	}
	return s.posts[idx], nil
}

func (s *inmemStorage) CreateComment(ctx context.Context, newComment model.NewComment) (*model.Comment, error) {
	comm := &model.Comment{
		ID:      uuid.New(),
		Author:  newComment.Author,
		Content: newComment.Content,
	}

	if newComment.PostID != nil {
		s.mu.Lock()
		defer s.mu.Unlock()

		idx := slices.IndexFunc(s.posts, func(post *model.Post) bool {
			return post.ID.String() == *newComment.PostID
		})
		if idx == -1 {
			return nil, ErrNotFound
		}
		if !s.posts[idx].Commentable {
			return nil, ErrNotCommentable
		}
		parsed, _ := uuid.Parse(*newComment.PostID)
		comm.PostID = &parsed
		s.posts[idx].Comments = append(s.posts[idx].Comments, comm)
	} else if newComment.CommentID != nil {
		s.mu.Lock()
		defer s.mu.Unlock()

		parentID := *newComment.CommentID
		for _, post := range s.posts {
			if insertComment(post.Comments, comm, parentID) {
				return comm, nil
			}
		}
		return nil, ErrNotFound
	} else {
		return nil, ErrBadRequest
	}

	return comm, nil
}

func insertComment(comments []*model.Comment, newComment *model.Comment, parentId string) bool {
	for _, comment := range comments {
		if comment.ID.String() == parentId {
			newComment.PostID = comment.PostID
			comment.Comments = append(comment.Comments, newComment)
			return true
		}

		if insertComment(comment.Comments, newComment, parentId) {
			return true
		}
	}

	return false
}
