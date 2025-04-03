package storage

import (
	"context"
	"time"
	"database/sql"
	"graphql_project/internal/graph/model"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresStorage_CreatePost(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &PostgresStorage{db: db}
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		newPost := model.NewPost{
			Title:       "Test Post",
			Author:      "Author",
			Content:     "Content",
			Commentable: true,
		}

		mock.ExpectExec("INSERT INTO posts").
			WithArgs(sqlmock.AnyArg(), newPost.Title, newPost.Author, newPost.Content, newPost.Commentable).
			WillReturnResult(sqlmock.NewResult(1, 1))

		post, err := storage.CreatePost(ctx, newPost)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, post.ID)
		assert.Equal(t, newPost.Title, post.Title)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		newPost := model.NewPost{Title: "Error Test"}
		mock.ExpectExec("INSERT INTO posts").
			WillReturnError(sql.ErrConnDone)

		_, err := storage.CreatePost(ctx, newPost)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresStorage_GetAllPosts(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &PostgresStorage{db: db}
	ctx := context.Background()

	postRows := sqlmock.NewRows([]string{"id", "title", "author", "content", "commentable"}).
		AddRow(uuid.New(), "Post 1", "Author", "Content", true).
		AddRow(uuid.New(), "Post 2", "Author", "Content", false)

	commentRows := sqlmock.NewRows([]string{"id", "post_id", "parent_comment_id", "author", "content"})

	t.Run("get all posts", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, title, author, content, commentable FROM posts").
			WillReturnRows(postRows)

		mock.ExpectQuery("SELECT id, post_id, parent_comment_id, author, content FROM comments").
			WillReturnRows(commentRows)

		posts, err := storage.GetAllPosts(ctx, nil, nil)
		require.NoError(t, err)
		assert.Len(t, posts, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresStorage_GetPostByID(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    storage := &PostgresStorage{db: db}
    ctx := context.Background()

    postID := uuid.New()
    commentID := uuid.New()
    nonExistentID := uuid.New().String()

    t.Run("existing post", func(t *testing.T) {
        mock.ExpectQuery("SELECT id, title, author, content, commentable FROM posts WHERE id = \\$1").
            WithArgs(postID.String()).
            WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "content", "commentable"}).
                AddRow(postID, "Test Post", "Author", "Content", true))

        mock.ExpectQuery("SELECT id, post_id, parent_comment_id, author, content FROM comments WHERE post_id = \\$1").
            WithArgs(postID).
            WillReturnRows(sqlmock.NewRows([]string{"id", "post_id", "parent_comment_id", "author", "content"}).
                AddRow(commentID, postID, nil, "User", "Comment"))

        post, err := storage.GetPostByID(ctx, postID.String())
        require.NoError(t, err)
        assert.Equal(t, postID, post.ID)
        assert.Len(t, post.Comments, 1)
        assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("not found", func(t *testing.T) {
        mock.ExpectQuery("SELECT id, title, author, content, commentable FROM posts WHERE id = \\$1").
            WithArgs(nonExistentID).
            WillReturnError(sql.ErrNoRows)

        _, err := storage.GetPostByID(ctx, nonExistentID)
        assert.ErrorIs(t, err, ErrNotFound)
        assert.NoError(t, mock.ExpectationsWereMet())
    })
}

func TestPostgresStorage_CreateComment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &PostgresStorage{db: db}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	postID := uuid.New()
	commentID := uuid.New()

	t.Run("comment to post", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT commentable FROM posts WHERE id = ?").
			WithArgs(postID).
			WillReturnRows(sqlmock.NewRows([]string{"commentable"}).AddRow(true))
		mock.ExpectExec("INSERT INTO comments").
			WithArgs(sqlmock.AnyArg(), postID, "Author", "Content").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		_, err := storage.CreateComment(ctx, model.NewComment{
			Author:  "Author",
			Content: "Content",
			PostID:  ptr(postID.String()),
		})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("comment to comment", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT post_id FROM comments WHERE id = ?").
			WithArgs(commentID).
			WillReturnRows(sqlmock.NewRows([]string{"post_id"}).AddRow(postID))
		mock.ExpectQuery("SELECT commentable FROM posts WHERE id = ?").
			WithArgs(postID).
			WillReturnRows(sqlmock.NewRows([]string{"commentable"}).AddRow(true))
		mock.ExpectExec("INSERT INTO comments").
			WithArgs(sqlmock.AnyArg(), postID, commentID, "Author", "Content").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		_, err := storage.CreateComment(ctx, model.NewComment{
			Author:     "Author",
			Content:    "Content",
			CommentID:  ptr(commentID.String()),
		})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("post not commentable", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT commentable FROM posts WHERE id = ?").
			WithArgs(postID).
			WillReturnRows(sqlmock.NewRows([]string{"commentable"}).AddRow(false))
		mock.ExpectRollback()

		_, err := storage.CreateComment(ctx, model.NewComment{
			Author:  "Author",
			Content: "Content",
			PostID:  ptr(postID.String()),
		})
		assert.ErrorIs(t, err, ErrNotCommentable)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func ptr(s string) *string { return &s }