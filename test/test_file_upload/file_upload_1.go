package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
)

func main() {
	file, err := os.Open(`./Gopher.png`) // 로컬 파일 경로 
	if err != nil {
		fmt.Printf("Unable to open file: %v", err)
		return
	}
	defer file.Close()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-2"), // S3 버킷의 리전
	})
	if err != nil {
		fmt.Printf("Unable to create session: %v", err)
		return
	}

	uploader := s3manager.NewUploader(sess)

	uuid := uuid.New() // UUID생성
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("kkamji-image-upload-test"), // S3 버킷 이름
		Key:    aws.String(uuid.String()),              // S3에 저장될 파일 이름 (uuid로)
		Body:   file,
	})

	if err != nil {
		fmt.Printf("Unable to upload file: %v", err)
		return
	}

	fmt.Println("File uploaded successfully")
}
