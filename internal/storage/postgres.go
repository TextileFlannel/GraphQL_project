package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"graphql_project/internal/graph/model"
	"strings"
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
	query := "SELECT id, title, author, content, commentable FROM posts"
	var args []interface{}

	if limit != nil {
		query += " LIMIT $1"
		args = append(args, *limit)
		if offset != nil {
			query += " OFFSET $2"
			args = append(args, *offset)
		}
	} else if offset != nil {
		query += " OFFSET $1"
		args = append(args, *offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		post.Comments = []*model.Comment{}
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return posts, nil
	}

	postIDs := make([]uuid.UUID, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}

	commentsQuery := fmt.Sprintf(
		"SELECT id, post_id, parent_comment_id, author, content FROM comments WHERE post_id IN (%s)",
		placeholders(len(postIDs)),
	)
	args = make([]interface{}, len(postIDs))
	for i, id := range postIDs {
		args[i] = id
	}

	rows, err = s.db.QueryContext(ctx, commentsQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %v", err)
	}
	defer rows.Close()

	var tempComments []model.TempComment
	for rows.Next() {
		var c model.TempComment
		if err := rows.Scan(&c.ID, &c.PostID, &c.ParentCommentID, &c.Author, &c.Content); err != nil {
			return nil, fmt.Errorf("scanning comment: %v", err)
		}
		tempComments = append(tempComments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("after scanning comments: %v", err)
	}

	commentMap := make(map[uuid.UUID]*model.Comment)
	commentsByPostID := make(map[uuid.UUID][]*model.Comment)

	for _, tc := range tempComments {
		comment := &model.Comment{
			ID:       tc.ID,
			Author:   tc.Author,
			Content:  tc.Content,
			PostID:   &tc.PostID,
			Comments: []*model.Comment{},
		}
		commentMap[tc.ID] = comment
	}

	for _, tc := range tempComments {
		comment := commentMap[tc.ID]
		if tc.ParentCommentID != nil {
			parent, exists := commentMap[*tc.ParentCommentID]
			if exists {
				parent.Comments = append(parent.Comments, comment)
			}
		} else {
			commentsByPostID[tc.PostID] = append(commentsByPostID[tc.PostID], comment)
		}
	}

	for _, post := range posts {
		if comments, ok := commentsByPostID[post.ID]; ok {
			post.Comments = comments
		} else {
			post.Comments = []*model.Comment{}
		}
	}

	return posts, nil
}

func (s *PostgresStorage) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	var post model.Post
	err := s.db.QueryRowContext(ctx,
		"SELECT id, title, author, content, commentable FROM posts WHERE id = $1",
		id,
	).Scan(
		&post.ID,
		&post.Title,
		&post.Author,
		&post.Content,
		&post.Commentable,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT id, post_id, parent_comment_id, author, content FROM comments WHERE post_id = $1",
		post.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %v", err)
	}
	defer rows.Close()

	var tempComments []model.TempComment
	for rows.Next() {
		var c model.TempComment
		if err := rows.Scan(&c.ID, &c.PostID, &c.ParentCommentID, &c.Author, &c.Content); err != nil {
			return nil, fmt.Errorf("scanning comment: %v", err)
		}
		tempComments = append(tempComments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("after scanning comments: %v", err)
	}

	commentMap := make(map[uuid.UUID]*model.Comment)
	commentsByPostID := make(map[uuid.UUID][]*model.Comment)

	for _, tc := range tempComments {
		comment := &model.Comment{
			ID:       tc.ID,
			Author:   tc.Author,
			Content:  tc.Content,
			PostID:   &tc.PostID,
			Comments: []*model.Comment{},
		}
		commentMap[tc.ID] = comment
	}

	for _, tc := range tempComments {
		comment := commentMap[tc.ID]
		if tc.ParentCommentID != nil {
			parent, exists := commentMap[*tc.ParentCommentID]
			if exists {
				parent.Comments = append(parent.Comments, comment)
			}
		} else {
			commentsByPostID[tc.PostID] = append(commentsByPostID[tc.PostID], comment)
		}
	}

	if comments, ok := commentsByPostID[post.ID]; ok {
		post.Comments = comments
	} else {
		post.Comments = []*model.Comment{}
	}

	return &post, nil
}

func (s *PostgresStorage) CreateComment(ctx context.Context, newComment model.NewComment) (*model.Comment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	comment := &model.Comment{
		ID:      uuid.New(),
		Author:  newComment.Author,
		Content: newComment.Content,
	}

	switch {
	case newComment.PostID != nil:
		postID, err := uuid.Parse(*newComment.PostID)
		if err != nil {
			return nil, ErrBadRequest
		}

		var commentable bool
		err = tx.QueryRowContext(ctx,
			"SELECT commentable FROM posts WHERE id = $1",
			postID,
		).Scan(&commentable)

		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		if !commentable {
			return nil, ErrNotCommentable
		}

		_, err = tx.ExecContext(ctx,
			"INSERT INTO comments (id, post_id, author, content) VALUES ($1, $2, $3, $4)",
			comment.ID, postID, comment.Author, comment.Content,
		)
		if err != nil {
			return nil, err
		}
		parsed, _ := uuid.Parse(*newComment.PostID)
		comment.PostID = &(parsed)

	case newComment.CommentID != nil:
		parentID, err := uuid.Parse(*newComment.CommentID)
		if err != nil {
			return nil, ErrBadRequest
		}

		var postID uuid.UUID
		err = tx.QueryRowContext(ctx,
			"SELECT post_id FROM comments WHERE id = $1",
			parentID,
		).Scan(&postID)
		comment.PostID = &postID

		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}

		var commentable bool
		err = tx.QueryRowContext(ctx,
			"SELECT commentable FROM posts WHERE id = $1",
			postID,
		).Scan(&commentable)

		if err != nil {
			return nil, err
		}
		if !commentable {
			return nil, ErrNotCommentable
		}

		_, err = tx.ExecContext(ctx,
			"INSERT INTO comments (id, post_id, parent_comment_id, author, content) VALUES ($1, $2, $3, $4, $5)",
			comment.ID, postID, parentID, comment.Author, comment.Content,
		)
		if err != nil {
			return nil, err
		}

	default:
		return nil, ErrBadRequest
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return comment, nil
}

func placeholders(n int) string {
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		parts[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(parts, ", ")
}
