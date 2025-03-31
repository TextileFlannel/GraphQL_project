package model

type Comment struct {
	ID       string              `json:"id"`
	PostID   string              `json:"postId"`
	ParentID *string             `json:"parentId,omitempty"`
	Content  string              `json:"content"`
	Author   string              `json:"author"`
	Replies  *CommentsConnection `json:"replies"`
}

type CommentEdge struct {
	Node   *Comment `json:"node"`
	Cursor string   `json:"cursor"`
}

type CommentsConnection struct {
	Edges    []*CommentEdge `json:"edges"`
	PageInfo *PageInfo      `json:"pageInfo"`
}

type Mutation struct {
}

type PageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor,omitempty"`
}

type Post struct {
	ID              string              `json:"id"`
	Title           string              `json:"title"`
	Content         string              `json:"content"`
	Author          string              `json:"author"`
	CommentsEnabled bool                `json:"commentsEnabled"`
	Comments        *CommentsConnection `json:"comments"`
}

type PostEdge struct {
	Node   *Post  `json:"node"`
	Cursor string `json:"cursor"`
}

type PostsConnection struct {
	Edges    []*PostEdge `json:"edges"`
	PageInfo *PageInfo   `json:"pageInfo"`
}

type Query struct {
}

type Subscription struct {
}
