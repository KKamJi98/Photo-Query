package picture

import(	
	"database/sql"
	"time"
	"encoding/json"
)

// Picture struct represents the structure of a picture record.
type Picture struct {
	PictureID  int64        `json:"picture_id"`
	UserID     int64        `json:"user_id"`
	ImageURL   string       `json:"image_url"`
	CreatedAt  CustomTime	`json:"created_at,omitempty"`
	DeletedAt  CustomTime	`json:"deleted_at,omitempty"`
	Bookmarked int8         `json:"bookmarked"`
}

type CustomTime struct {
	sql.NullTime
}

// MarshalJSON is a custom marshaller that omits the Valid field and formats the Time.
// func (ct *CustomTime) MarshalJSON() ([]byte, error) {
// 	if !ct.Valid {
// 		return []byte("null"), nil
// 	}
// 	// You can also format the time as you wish here.
// 	return json.Marshal(ct.Time.Format(time.RFC3339))
// }

// // UnmarshalJSON is a custom unmarshaller that parses time and sets Valid field accordingly.
// func (ct *CustomTime) UnmarshalJSON(data []byte) error {
// 	// If the data is "null", set Valid to false.
// 	if string(data) == "null" {
// 		ct.Valid = false
// 		return nil
// 	}
// 	// Assume the input is in RFC3339 format for parsing.
// 	t, err := time.Parse(`"`+time.RFC3339+`"`, string(data))
// 	if err != nil {
// 		return err
// 	}
// 	ct.Valid = true
// 	ct.Time = t
// 	return nil
// }