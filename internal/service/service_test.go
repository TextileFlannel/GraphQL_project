package service

import (
	"context"
	"graphql_project/internal/graph/model"
	"graphql_project/internal/storage"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) CreatePost(ctx context.Context, newPost model.NewPost) (*model.Post, error) {
	args := m.Called(ctx, newPost)
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockStorage) GetAllPosts(ctx context.Context, offset *int, limit *int) ([]*model.Post, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*model.Post), args.Error(1)
}

func (m *MockStorage) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockStorage) CreateComment(ctx context.Context, newComment model.NewComment) (*model.Comment, error) {
	args := m.Called(ctx, newComment)
	return args.Get(0).(*model.Comment), args.Error(1)
}

func TestService_CreatePost(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MockStorage)
	service := NewService(mockStorage)

	newPost := model.NewPost{
		Title:       "Test Post",
		Author:      "Author",
		Content:     "Content",
		Commentable: true,
	}

	expectedPost := &model.Post{
		ID:          uuid.New(),
		Title:       newPost.Title,
		Author:      newPost.Author,
		Content:     newPost.Content,
		Commentable: newPost.Commentable,
	}

	t.Run("success", func(t *testing.T) {
		mockStorage.On("CreatePost", ctx, newPost).
			Return(expectedPost, nil).
			Once()

		result, err := service.CreatePost(ctx, newPost)

		require.NoError(t, err)
		assert.Equal(t, expectedPost, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage error", func(t *testing.T) {
		mockStorage.On("CreatePost", ctx, newPost).
			Return((*model.Post)(nil), storage.ErrNotFound).
			Once()

		_, err := service.CreatePost(ctx, newPost)

		assert.ErrorIs(t, err, storage.ErrNotFound)
		mockStorage.AssertExpectations(t)
	})
}

func TestService_GetAllPosts(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MockStorage)
	service := NewService(mockStorage)

	posts := []*model.Post{
		{ID: uuid.New(), Title: "Post 1"},
		{ID: uuid.New(), Title: "Post 2"},
	}

	t.Run("success without pagination", func(t *testing.T) {
		mockStorage.On("GetAllPosts", ctx, (*int)(nil), (*int)(nil)).
			Return(posts, nil).
			Once()

		result, err := service.GetAllPosts(ctx, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, posts, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("success with pagination", func(t *testing.T) {
		offset := 2
		limit := 5
		mockStorage.On("GetAllPosts", ctx, &offset, &limit).
			Return(posts, nil).
			Once()

		result, err := service.GetAllPosts(ctx, &offset, &limit)

		require.NoError(t, err)
		assert.Equal(t, posts, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage error", func(t *testing.T) {
		mockStorage.On("GetAllPosts", ctx, (*int)(nil), (*int)(nil)).
			Return([]*model.Post(nil), storage.ErrNotFound).
			Once()

		_, err := service.GetAllPosts(ctx, nil, nil)

		assert.ErrorIs(t, err, storage.ErrNotFound)
		mockStorage.AssertExpectations(t)
	})
}

func TestService_GetPostByID(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MockStorage)
	service := NewService(mockStorage)

	postID := uuid.New().String()
	post := &model.Post{ID: uuid.MustParse(postID), Title: "Test Post"}

	t.Run("success", func(t *testing.T) {
		mockStorage.On("GetPostByID", ctx, postID).
			Return(post, nil).
			Once()

		result, err := service.GetPostByID(ctx, postID)

		require.NoError(t, err)
		assert.Equal(t, post, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockStorage.On("GetPostByID", ctx, postID).
			Return((*model.Post)(nil), storage.ErrNotFound).
			Once()

		_, err := service.GetPostByID(ctx, postID)

		assert.ErrorIs(t, err, storage.ErrNotFound)
		mockStorage.AssertExpectations(t)
	})
}

func TestService_CreateComment(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	mockStorage := new(MockStorage)
	service := NewService(mockStorage)

	postID := uuid.New().String()
	commentID := uuid.New().String()
	newComment := model.NewComment{
		Author:  "User",
		Content: "Comment",
		PostID:  &postID,
	}

	expectedComment := &model.Comment{
		ID:      uuid.MustParse(commentID),
		Author:  newComment.Author,
		Content: newComment.Content,
	}

	t.Run("success to post", func(t *testing.T) {
		mockStorage.On("CreateComment", ctx, newComment).
			Return(expectedComment, nil).
			Once()

		result, err := service.CreateComment(ctx, newComment)

		require.NoError(t, err)
		assert.Equal(t, expectedComment, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("success to comment", func(t *testing.T) {
		commentID := uuid.New().String()
		newComment := model.NewComment{
			Author:    "User",
			Content:   "Reply",
			CommentID: &commentID,
		}

		mockStorage.On("CreateComment", ctx, newComment).
			Return(expectedComment, nil).
			Once()

		result, err := service.CreateComment(ctx, newComment)

		require.NoError(t, err)
		assert.Equal(t, expectedComment, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage error", func(t *testing.T) {
		mockStorage.On("CreateComment", ctx, newComment).
			Return((*model.Comment)(nil), storage.ErrNotCommentable).
			Once()

		_, err := service.CreateComment(ctx, newComment)

		assert.ErrorIs(t, err, storage.ErrNotCommentable)
		mockStorage.AssertExpectations(t)
	})

	t.Run("invalid request", func(t *testing.T) {
		invalidComment := model.NewComment{
			Author:  "User",
			Content: "Comment",
		}

		mockStorage.On("CreateComment", ctx, invalidComment).
			Return((*model.Comment)(nil), storage.ErrBadRequest).
			Once()

		_, err := service.CreateComment(ctx, invalidComment)

		assert.ErrorIs(t, err, storage.ErrBadRequest)
		mockStorage.AssertExpectations(t)
	})
}
