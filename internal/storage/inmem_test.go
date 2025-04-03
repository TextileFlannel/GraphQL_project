package storage

import (
	"context"
	"graphql_project/internal/graph/model"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePost(t *testing.T) {
	s := NewInMemStorage()
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		newPost := model.NewPost{
			Title:       "Test Post",
			Author:      "Author",
			Content:     "Content",
			Commentable: true,
		}

		post, err := s.CreatePost(ctx, newPost)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, post.ID)
		assert.Equal(t, newPost.Title, post.Title)
		assert.Equal(t, newPost.Author, post.Author)
		assert.Equal(t, newPost.Content, post.Content)
		assert.Equal(t, 1, len(s.posts))
	})
}

func TestGetAllPosts(t *testing.T) {
	s := NewInMemStorage()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		s.posts = append(s.posts, &model.Post{
			ID:          uuid.New(),
			Title:       "Post",
			Author:      "Author",
			Content:     "Content",
			Commentable: true,
		})
	}

	t.Run("get all posts", func(t *testing.T) {
		posts, err := s.GetAllPosts(ctx, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, 5, len(posts))
	})
}

func TestGetPostByID(t *testing.T) {
	s := NewInMemStorage()
	ctx := context.Background()

	post := &model.Post{
		ID:          uuid.New(),
		Title:       "Test Post",
		Author:      "Author",
		Content:     "Content",
		Commentable: true,
	}
	s.posts = append(s.posts, post)

	t.Run("existing post", func(t *testing.T) {
		found, err := s.GetPostByID(ctx, post.ID.String())
		require.NoError(t, err)
		assert.Equal(t, post.ID, found.ID)
	})

	t.Run("non-existent post", func(t *testing.T) {
		_, err := s.GetPostByID(ctx, uuid.NewString())
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestCreateComment(t *testing.T) {
	s := NewInMemStorage()
	ctx := context.Background()

	post, err := s.CreatePost(ctx, model.NewPost{
		Title:       "Test Post",
		Author:      "Author",
		Content:     "Content",
		Commentable: true,
	})
	require.NoError(t, err)

	postIDStr := post.ID.String()
	comment, err := s.CreateComment(ctx, model.NewComment{
		Author:  "Commenter",
		Content: "Test Comment",
		PostID:  &postIDStr,
	})
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, comment.ID)
	assert.Equal(t, "Commenter", comment.Author)
	assert.Equal(t, "Test Comment", comment.Content)
	assert.Equal(t, post.ID, *comment.PostID)

	postFromStorage, err := s.GetPostByID(ctx, post.ID.String())
	require.NoError(t, err)

	if assert.Len(t, postFromStorage.Comments, 1) {
		assert.Equal(t, comment.ID, postFromStorage.Comments[0].ID)
		assert.Equal(t, "Test Comment", postFromStorage.Comments[0].Content)
	}
}
