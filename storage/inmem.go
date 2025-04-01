package storage

import (
	"context"
	"errors"
	"slices"
	"sync"

	"graphql_project/graph/model"
	"github.com/google/uuid"
)

var (
	ErrNotCommentable = errors.New("the post is not commentable")
	ErrNotFound       = errors.New("not found")
	ErrBadRequest     = errors.New("bad request")
)

type inmemStorage struct {
	posts []*model.Post
}

func NewInMemStorage() *inmemStorage {
	return &inmemStorage{make([]*model.Post, 0)}
}

func (s *inmemStorage) CreatePost(ctx context.Context, newPost model.NewPost) (*model.Post, error) {
	post := &model.Post{
		ID:          uuid.New(),
		Title:       newPost.Title,
		Author:      newPost.Author,
		Content:     newPost.Content,
		Commentable: newPost.Commentable,
	}
	s.posts = append(s.posts, post)
	return post, nil
}

func (s *inmemStorage) GetAllPosts(ctx context.Context, offset *int, limit *int) ([]*model.Post, error) {
	off := 0
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
		idx := slices.IndexFunc(s.posts, func(post *model.Post) bool {
			return post.ID.String() == *newComment.PostID
		})
		if idx == -1 {
			return nil, ErrNotFound
		}
		if !s.posts[idx].Commentable {
			return nil, ErrNotCommentable
		}
		s.posts[idx].Comments = append(s.posts[idx].Comments, comm)
	} else if newComment.CommentID != nil {
		var wg sync.WaitGroup
		for _, post := range s.posts {
			wg.Add(1)
			go func(p *model.Post) {
				defer wg.Done()
				insertComment(p.Comments, comm, *newComment.CommentID)
			}(post)
		}
		wg.Wait()
	} else {
		return nil, ErrBadRequest
	}

	return comm, nil
}

func insertComment(comments []*model.Comment, newComment *model.Comment, parentId string) {
	if comments == nil {
		return
	}

	for _, comment := range comments {
		if comment.ID.String() == parentId {
			comment.Comments = append(comment.Comments, newComment)
			return
		} else {
			go insertComment(comment.Comments, newComment, parentId)
		}
	}
}