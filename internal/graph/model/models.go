package model

import (
	"github.com/google/uuid"
)

type TempComment struct {
	ID              uuid.UUID
	PostID          uuid.UUID
	ParentCommentID *uuid.UUID
	Author          string
	Content         string
}
