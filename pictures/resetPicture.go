package picture

import (
	"ace-app/databases"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gin-gonic/gin"
)

// 그림들을 ID로 삭제하는 작업을 처리
func DeleteAllPictures(c *gin.Context) {
	// 요청에서 JSON 데이터를 추출하고 언마샬합니다.
	db := database.ConnectDB()
	defer db.Close()

	// s3에서 이미지파일 삭제
	var pictures []Picture
	userId := c.Query("user_id")
	log.Println(userId)
	rows, err := db.Query("SELECT user_id, image_url FROM Pictures WHERE user_id = ?", userId)
	if err != nil {
		log.Printf("사용자 %v에 대한 사진 조회 오류: %v", userId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.UserID, &picture.ImageURL); err != nil {
			log.Printf("사진 스캔 오류: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}
	log.Println(len(pictures))

	// user_id로 삭제 시도
	result, err := db.Exec("DELETE FROM Pictures WHERE user_id = ?", userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "사진 삭제 실패"})
	}
	rowsAffected, err := result.RowsAffected()
	s3DeleteAllPictures(c, pictures, db)
	deleteItemsByPartitionKey(os.Getenv("DYNAMODB_TABLE"), "user_id", userId)
	log.Printf("%d개 사진 삭제 완료", rowsAffected)
	// 삭제 작업 결과 응답
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d개 사진 삭제 완료", rowsAffected)})
}

func s3DeleteAllPictures(c *gin.Context, pictures []Picture, db *sql.DB) {
	var s3ImageOriginalObjectKeys []string
	var s3ImageThumbnailObjectKeys []string

	log.Println(len(pictures))
	for _, pic := range pictures {
		log.Printf("ImageURL => %v \t PictureID => %v", pic.ImageURL, pic.UserID)
		s3ImageOriginalObjectKeys = append(s3ImageOriginalObjectKeys, fmt.Sprintf("%s/%s/%s", "original", pic.UserID, pic.ImageURL))
		s3ImageThumbnailObjectKeys = append(s3ImageThumbnailObjectKeys, fmt.Sprintf("%s/%s/%s", "thumbnail", pic.UserID, pic.ImageURL))
	}

	// S3 클라이언트 초기화
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	s3Client := s3.NewFromConfig(cfg)

	// BucketBasics 인스턴스 생성
	basics := BucketBasics{S3Client: s3Client}
	basics.DeleteObjects(c, "rapa-app-image-bucket", s3ImageOriginalObjectKeys)
	basics.DeleteObjects(c, "rapa-app-image-bucket", s3ImageThumbnailObjectKeys)
}

func deleteItemsByPartitionKey(tableName, partitionKeyName, partitionKeyValue string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("AWS config 로딩 실패: %v", err)
		return err
	}
	svc := dynamodb.NewFromConfig(cfg)

	// 파티션 키로 모든 항목 쿼리
	queryOutput, err := svc.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              &tableName,
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :v", partitionKeyName)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v": &types.AttributeValueMemberS{Value: partitionKeyValue},
		},
	})
	if err != nil {
		log.Printf("항목 쿼리 실패: %v", err)
		return err
	}

	// 쿼리 결과로 얻은 각 항목 삭제
	for _, item := range queryOutput.Items {
		sortKeyValue, ok := item["image_url"].(*types.AttributeValueMemberS)
		if !ok {
			log.Println("정렬 키(image_url) 추출 실패 또는 키 타입 불일치")
			continue // 다음 항목 처리
		}

		_, err := svc.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
			TableName: &tableName,
			Key: map[string]types.AttributeValue{
				partitionKeyName: &types.AttributeValueMemberS{Value: partitionKeyValue},
				"image_url":      sortKeyValue, // 정렬 키 이름을 "SortKeyName"에서 "image_url"로 변경
			},
		})
		if err != nil {
			log.Printf("항목 삭제 실패: %v", err)
			// 실패한 항목에 대해 로깅하고 계속 진행할 수 있음
			// 여기서는 에러를 반환하여 중단
			return err
		}
	}

	return nil
}
