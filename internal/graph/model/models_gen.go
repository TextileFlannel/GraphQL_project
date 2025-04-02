// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"github.com/google/uuid"
)

type Comment struct {
	ID       uuid.UUID  `json:"id"`
	Author   string     `json:"author"`
	Content  string     `json:"content"`
	PostID   *uuid.UUID `json:"postId,omitempty"`
	Comments []*Comment `json:"comments,omitempty"`
}

type Mutation struct {
}

type NewComment struct {
	Content   string  `json:"content"`
	Author    string  `json:"author"`
	CommentID *string `json:"commentId,omitempty"`
	PostID    *string `json:"postId,omitempty"`
}

type NewPost struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	Commentable bool   `json:"commentable"`
	Author      string `json:"author"`
}

type Post struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Author      string     `json:"author"`
	Content     string     `json:"content"`
	Commentable bool       `json:"commentable"`
	Comments    []*Comment `json:"comments,omitempty"`
}

type Query struct {
}

type Subscription struct {
}
