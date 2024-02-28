package picture

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Picture 구조체는 사진 레코드의 구조를 나타냅니다.
type Picture struct {
	PictureID  int64      `json:"picture_id"`           // 사진 ID
	UserID     string     `json:"user_id"`              // 사용자 ID
	ImageURL   string     `json:"image_url"`            // 이미지 URL
	CreatedAt  CustomTime `json:"created_at,omitempty"` // 생성 시간, 값이 없으면 JSON에서 생략
	Bookmarked int8       `json:"bookmarked"`           // 북마크 여부
}

// CustomTime 구조체 => sql.NullTime을 확장 => JSON 마샬링 및 언마샬링 커스텀
type CustomTime struct {
	sql.NullTime // Null 가능한 시간 값
}

// 커스텀 마샬러로, Valid 필드를 생략하고 시간을 포맷
func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	if !ct.Valid {
		return []byte("null"), nil // Valid가 false이면 null을 반환
	}
	// Valid가 true이면 RFC3339 포맷으로 시간을 반환
	return json.Marshal(ct.Time.Format(time.RFC3339))
}

// 커스텀 언마샬러, 시간 파싱하고 Valid 필드 설정
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	// 데이터가 "null"이면, Valid를 false로 설정
	if string(data) == "null" {
		ct.Valid = false
		return nil
	}
	// 데이터를 RFC3339 포맷으로 파싱
	t, err := time.Parse(`"`+time.RFC3339+`"`, string(data))
	if err != nil {
		return err // 파싱 에러 발생 시 반환
	}
	ct.Valid = true // 파싱 성공 시, Valid를 true로 설정
	ct.Time = t
	return nil
}

// Tag 구조체 정의
type Tag struct {
	UserId  string   `dynamodbav:"user_id"`
	ImageId string   `dynamodbav:"image_url"`
	Tags    []string `dynamodbav:"tags"`
	UploadTime int64 `dynamodbav:"upload_time"`
}

type TableBasics struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}
