package picture

import(	
	"database/sql"
)

// Picture struct represents the structure of a picture record.
type Picture struct {
	PictureID  int64        `json:"picture_id"`
	UserID     int64        `json:"user_id"`
	ImageURL   string       `json:"image_url"`
	CreatedAt  sql.NullTime `json:"created_at"`
	DeletedAt  sql.NullTime `json:"deleted_at"`
	Bookmarked int8         `json:"bookmarked"`
}