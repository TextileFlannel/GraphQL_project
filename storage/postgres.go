package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"graphql_project/graph/model"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) CreatePost(ctx context.Context, newPost model.NewPost) (*model.Post, error) {
	post := &model.Post{
		ID:          uuid.New(),
		Title:       newPost.Title,
		Author:      newPost.Author,
		Content:     newPost.Content,
		Commentable: newPost.Commentable,
		Comments:    []*model.Comment{},
	}

	_, err := s.db.ExecContext(ctx,
		"INSERT INTO posts(id, title, author, content, commentable) VALUES($1, $2, $3, $4, $5)",
		post.ID, post.Title, post.Author, post.Content, post.Commentable,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create post: %v", err)
	}

	return post, nil
}

func (s *PostgresStorage) GetAllPosts(ctx context.Context, offset *int, limit *int) ([]*model.Post, error) {
	defaultLimit := 10
	if limit != nil && *limit <= 0 {
		return nil, ErrBadRequest
	}

	query := "SELECT id, title, author, content, commentable FROM posts"
	args := []interface{}{}

	if limit != nil {
		query += " LIMIT $1"
		args = append(args, *limit)
		if offset != nil {
			query += " OFFSET $2"
			args = append(args, *offset)
		}
	} else {
		query += " LIMIT $1"
		args = append(args, defaultLimit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %v", err)
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Author,
			&post.Content,
			&post.Commentable,
		); err != nil {
			return nil, fmt.Errorf("failed to scan post: %v", err)
		}

		comments, err := s.getCommentsByPostID(ctx, post.ID)
		if err != nil {
			return nil, err
		}
		post.Comments = comments

		posts = append(posts, &post)
	}

	return posts, nil
}

func (s *PostgresStorage) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	postID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}

	var post model.Post
	err = s.db.QueryRowContext(ctx,
		"SELECT id, title, author, content, commentable FROM posts WHERE id = $1",
		postID,
	).Scan(
		&post.ID,
		&post.Title,
		&post.Author,
		&post.Content,
		&post.Commentable,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get post: %v", err)
	}

	comments, err := s.getCommentsByPostID(ctx, post.ID)
	if err != nil {
		return nil, err
	}
	post.Comments = comments

	return &post, nil
}

func (s *PostgresStorage) CreateComment(ctx context.Context, newComment model.NewComment) (*model.Comment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	comment := &model.Comment{
		ID:      uuid.New(),
		Author:  newComment.Author,
		Content: newComment.Content,
	}

	if newComment.PostID != nil {
		postID, err := uuid.Parse(*newComment.PostID)
		if err != nil {
			return nil, ErrBadRequest
		}

		var commentable bool
		err = tx.QueryRowContext(ctx,
			"SELECT commentable FROM posts WHERE id = $1",
			postID,
		).Scan(&commentable)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("failed to check commentable: %v", err)
		}

		if !commentable {
			return nil, ErrNotCommentable
		}

		comment.PostID = &postID
		_, err = tx.ExecContext(ctx,
			"INSERT INTO comments(id, post_id, author, content) VALUES($1, $2, $3, $4)",
			comment.ID, postID, comment.Author, comment.Content,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert comment: %v", err)
		}

	} else if newComment.CommentID != nil {
		parentID, err := uuid.Parse(*newComment.CommentID)
		if err != nil {
			return nil, ErrBadRequest
		}

		var postID uuid.UUID
		err = tx.QueryRowContext(ctx,
			"SELECT post_id FROM comments WHERE id = $1",
			parentID,
		).Scan(&postID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("failed to get parent comment: %v", err)
		}

		var commentable bool
		err = tx.QueryRowContext(ctx,
			"SELECT commentable FROM posts WHERE id = $1",
			postID,
		).Scan(&commentable)

		if err != nil {
			return nil, fmt.Errorf("failed to check commentable: %v", err)
		}

		if !commentable {
			return nil, ErrNotCommentable
		}

		comment.PostID = &postID
		_, err = tx.ExecContext(ctx,
			"INSERT INTO comments(id, post_id, parent_comment_id, author, content) VALUES($1, $2, $3, $4, $5)",
			comment.ID, postID, parentID, comment.Author, comment.Content,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert nested comment: %v", err)
		}

	} else {
		return nil, ErrBadRequest
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return comment, nil
}

func (s *PostgresStorage) getCommentsByPostID(ctx context.Context, postID uuid.UUID) ([]*model.Comment, error) {
	rows, err := s.db.QueryContext(ctx,
		`WITH RECURSIVE comment_tree AS (
			SELECT id, author, content, post_id, parent_comment_id
			FROM comments
			WHERE post_id = $1 AND parent_comment_id IS NULL
			UNION ALL
			SELECT c.id, c.author, c.content, c.post_id, c.parent_comment_id
			FROM comments c
			INNER JOIN comment_tree ct ON ct.id = c.parent_comment_id
		)
		SELECT id, author, content, parent_comment_id FROM comment_tree`,
		postID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %v", err)
	}
	defer rows.Close()

	comments := make(map[uuid.UUID]*model.Comment)
	var rootComments []*model.Comment

	for rows.Next() {
		var comment model.Comment
		var parentID *uuid.UUID
		err := rows.Scan(
			&comment.ID,
			&comment.Author,
			&comment.Content,
			&parentID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %v", err)
		}

		comment.PostID = &postID
		comments[comment.ID] = &comment

		if parentID == nil {
			rootComments = append(rootComments, &comment)
		} else {
			parent := comments[*parentID]
			parent.Comments = append(parent.Comments, &comment)
		}
	}

	return rootComments, nil
}
