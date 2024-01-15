package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	r := gin.Default()

	// 파일 업로드 엔드포인트
	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file") // 프론트엔드에서 전송된 파일 수신
		if err != nil {
			c.JSON(500, gin.H{"message": "File upload error"})
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{"message": "File open error"})
			return
		}
		defer src.Close()

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("ap-northeast-2"), // S3 버킷의 리전
		})
		if err != nil {
			c.JSON(500, gin.H{"message": "AWS session error"})
			return
		}

		uploader := s3manager.NewUploader(sess)
		uuid := uuid.New()
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String("kkamji-image-upload-test"), // S3 버킷 이름
			Key:    aws.String(uuid.String()),              // S3에 저장될 파일 이름 (uuid로)
			Body:   src,
		})

		if err != nil {
			c.JSON(500, gin.H{"message": fmt.Sprintf("Unable to upload file: %v", err)})
			return
		}

		c.JSON(200, gin.H{"message": "File uploaded successfully"})
	})

	// 서버 시작
	r.Run(":8080")
}
