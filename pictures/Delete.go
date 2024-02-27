package picture

import (
	"ace-app/databases"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
)

// DeletePicturesByPostId는 그림들을 ID로 삭제하는 작업을 처리합니다.
func DeletePicturesByPostId(c *gin.Context) {

	// 예상되는 JSON 데이터 형식과 일치하는 구조체를 정의합니다.
	// type picture struct {
	// 	PictureID int64 `json:"picture_id"`
	// }

	var pictures []Picture

	// 요청에서 JSON 데이터를 추출하고 언마샬합니다.
	jsonData := c.PostForm("json_data")
	log.Printf("%v \t %T", jsonData, jsonData)
	if err := json.Unmarshal([]byte(jsonData), &pictures); err != nil {
		log.Printf("잘못된 JSON 데이터: %v", err)
		c.JSON(400, gin.H{"message": "잘못된 JSON 데이터", "error": err.Error()})
		return
	}

	db := database.ConnectDB()
	defer db.Close()

	// s3에서 이미지파일 삭제
	s3DeletePictures(c, pictures, db)

	// 삭제 작업 결과를 추적하기 위한 카운터를 초기화합니다.
	errorCount := 0
	successCount := 0
	nofoundCount := 0

	// 각 그림 ID를 반복하고 삭제를 시도합니다.
	for _, pic := range pictures {
		//! 데이터베이스에서 사진 삭제
		result, err := db.Exec("DELETE FROM Pictures WHERE picture_id = ?", pic.PictureID)
		if err != nil {
			log.Printf("%d번 ID를 가진 그림 삭제 중 오류 발생: %v", pic.PictureID, err)
			errorCount++
			continue
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("%d번 ID에 대한 영향 받는 행 수를 가져오는 중 오류 발생: %v", pic.PictureID, err)
			errorCount++
			continue
		}

		if rowsAffected == 0 {
			log.Printf("%d번 ID를 가진 그림을 찾을 수 없음", pic.PictureID)
			nofoundCount++
		} else {
			successCount++
		}
	}

	// 작업 결과를 로그에 기록합니다.
	log.Printf("%d개의 그림 삭제 || %d개의 오류 발생 || %d개의 그림 ID를 찾을 수 없음", successCount, errorCount, nofoundCount)

	// 삭제 작업 결과를 응답합니다.
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d개의 그림 삭제, %d개의 오류 발생, %d개의 그림 ID를 찾을 수 없음", successCount, errorCount, nofoundCount)})
}

func s3DeletePictures(c *gin.Context, picures []Picture, db *sql.DB) {
	// db := database.ConnectDB()
	// defer db.Close()
	// TODO
	//* picutre_id를 사용해서 이미지 이름, user_id 찾기
	//* 이것을 키값으로 S3 Bucket에서 이미지 삭제
	var imageName, userId string
	var s3ImageOriginalObjectKeys []string
	var s3ImageThumbnailObjectKeys []string
	var pictureNamesToDelete []string

	for _, pic := range picures {
		query := db.QueryRow("SELECT image_url, user_id FROM Pictures WHERE picture_id =?", pic.PictureID)
		err := query.Scan(&imageName, &userId)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("No results found.")
			} else {
				log.Printf("Query failed: %v", err)
			}
		}
		s3ImageOriginalObjectKeys = append(s3ImageOriginalObjectKeys, fmt.Sprintf("%s/%s/%s", "original", userId, imageName))
		s3ImageThumbnailObjectKeys = append(s3ImageThumbnailObjectKeys, fmt.Sprintf("%s/%s/%s", "thumbnail", userId, imageName))

		pictureNamesToDelete = append(pictureNamesToDelete, imageName)
	}
	// log.Printf("userId => %v\t imageName => %v \t %T", userId, imageName)

	// log.Printf("%v", s3ImageOriginalObjectKeys)

	// S3 클라이언트 초기화 (예시 코드, 실제 구현은 환경에 맞게 조정)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	s3Client := s3.NewFromConfig(cfg)

	// BucketBasics 인스턴스 생성
	basics := BucketBasics{S3Client: s3Client}
	basics.DeleteObjects(c, "rapa-app-image-bucket", s3ImageOriginalObjectKeys)
	basics.DeleteObjects(c, "rapa-app-image-bucket", s3ImageThumbnailObjectKeys)

	dynamoDbClient := dynamodb.NewFromConfig(cfg)
	tableBasics := TableBasics{
		DynamoDbClient: dynamoDbClient,
		TableName:      os.Getenv("DYNAMODB_TABLE"),
	}
	DeleteDynamoDBPictures(c, pictureNamesToDelete, tableBasics)
}

type BucketBasics struct {
	S3Client *s3.Client
}

func (basics BucketBasics) DeleteObjects(c *gin.Context, bucketName string, objectKeys []string) error {
	if len(objectKeys) == 0 {
		log.Printf("삭제할 객체 키가 제공되지 않았습니다.")
		return nil // 또는 적절한 에러 반환
	}

	var objectIdentifiers []types.ObjectIdentifier
	for _, key := range objectKeys {
		log.Println(key)
		objectIdentifiers = append(objectIdentifiers, types.ObjectIdentifier{Key: aws.String(key)})
	}

	output, err := basics.S3Client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &types.Delete{Objects: objectIdentifiers},
	})
	if err != nil {
		log.Printf("%v 버킷에서 객체를 삭제하지 못했습니다. 원인: %v\n", bucketName, err)
		return err
	}

	// log.Printf("%v",output)
	if len(output.Errors) > 0 {
		for _, e := range output.Errors {
			log.Printf("객체 %s 삭제 에러: %s", *e.Key, *e.Message)
		}
	} else {
		log.Printf("%v 개의 객체를 삭제했습니다.\n", len(output.Deleted))
	}

	return nil
}

func DeleteDynamoDBPictures(c *gin.Context, pictures []string, basics TableBasics) {
	count := 0
	for _, pictureId := range pictures {
		if count % 100 == 0 {
			time.Sleep(time.Second * 1)
		}
		err := basics.DeleteDynamoDBPicture(c, pictureId)
		if err != nil {
			log.Printf("Error deleting picture with ID %s: %v", pictureId, err)
		}
		count++
	}
}

func (basics TableBasics) DeleteDynamoDBPicture(c *gin.Context, pictureId string) error {
	userId := c.Param("user_id")
	key := map[string]dynamodbTypes.AttributeValue{
		"user_id":   &dynamodbTypes.AttributeValueMemberS{Value: userId},
		"image_url": &dynamodbTypes.AttributeValueMemberS{Value: pictureId},
	}

	_, err := basics.DynamoDbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(basics.TableName),
		Key:       key,
	})
	if err != nil {
		log.Printf("Couldn't delete %v from the table. Here's why: %v\n", pictureId, err)
		return err
	}
	return nil
}
