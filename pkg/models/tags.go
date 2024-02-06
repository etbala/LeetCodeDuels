package models

// Tag represents a tag entity as stored in your database.
type Tag struct {
	TagID   int    `json:"tag_id"`
	TagName string `json:"tag_name"`
}

// NewTag creates a new instance of Tag.
func NewTag(tagID int, tagName string) *Tag {
	return &Tag{
		TagID:   tagID,
		TagName: tagName,
	}
}
