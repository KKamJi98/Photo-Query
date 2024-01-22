package post

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

func CreatePost(c *gin.Context) {
	err := godotenv.Load("./env/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbEndpoint := os.Getenv("DB_ENDPOINT")
	dbName := os.Getenv("DB_NAME")

	// 데이터베이스 연결
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", dbUser, dbPassword, dbEndpoint, dbName))
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	defer db.Close()

	// 데이터베이스 연결 테스트
	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to database: ", err)
	}

	jsonData := c.PostForm("json_data") // "json_data"는 프론트엔드에서 전송하는 JSON 데이터의 필드 이름
	// json_data를 post 구조체의 변수에 적용
	var post struct {
		UserId  int64  `json:"user_id"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(jsonData), &post); err != nil {
		c.JSON(400, gin.H{"message": "Invalid JSON data", "error": err.Error()})
		return
	}

	file, err := c.FormFile("image")
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
		Region: aws.String("ap-northeast-2"), //S3 Bucket의 Region
	})
	if err != nil {
		c.JSON(500, gin.H{"message": "AWS session error"})
		return
	}

	uploader := s3manager.NewUploader(sess)
	uuid := uuid.New()
	uploadOutput, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("kkamji-image-upload-test"),
		Key:    aws.String(uuid.String()),
		Body:   src,
	})

	if err != nil {
		c.JSON(500, gin.H{"message": fmt.Sprintf("Unable to upload file: %v", err)})
		return
	}

	imageURL := uploadOutput.Location // S3에 업로드된 이미지 URL
	log.Printf("S3 image name => %v", imageURL)
	currentTime := time.Now()

	// RDS에 데이터 저장
	_, err = db.Exec("INSERT INTO posts (user_id, image_url, content, create_at, update_at) VALUES (?, ?, ?, ?, ?)",
		post.UserId, imageURL, post.Content, currentTime, currentTime)
	if err != nil {
		log.Printf("post.UserId => %v", post.UserId)   // UserId 확인용 코드
		log.Printf("post.Content => %v", post.Content) // Content 확인용 코드
		c.JSON(500, gin.H{"message": fmt.Sprintf("Unable to save post data: %v", err)})
		return
	}

	c.JSON(200, gin.H{"message": "File and post data uploaded successfully"})
}
