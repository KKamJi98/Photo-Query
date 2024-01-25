package post

import (
	"archive/zip"
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
	"strings"
	// "time"
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

	form, _ := c.MultipartForm()
	fileHeader := form.File["images"]
	if err != nil {
		c.JSON(500, gin.H{"message": "File upload error"})
		return
	}

	//aws session 생성
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-2"), //S3 Bucket의 Region
	})
	if err != nil {
		c.JSON(500, gin.H{"message": "AWS session error"})
		return
	}

	uploadFileCount := 0
	for _, file := range fileHeader {
		// fmt.Printf("file => %T\t fileHeader => %v\n", file, fileHeader)
		// log.Println(file.Filename)

		src, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{"message": "File open error"})
			return
		}
		defer src.Close()

		//만약 zip파일일 때
		if strings.HasSuffix(file.Filename, ".zip") {
			// fileInfo, _ := file.Stat()
			// ZIP 파일 처리
			zipReader, err := zip.NewReader(src, file.Size)
			if err != nil {
				c.JSON(500, gin.H{"message": "Error reading zip file"})
				return
			}

			// ZIP 파일 내의 파일들을 순회
			for _, imageFile := range zipReader.File {
				// ZIP 내의 파일 열기
				zipFileReader, err := imageFile.Open()
				if err != nil {
					continue // 다음 파일로 넘어감
				}
				defer zipFileReader.Close()
				
				// ZIP 파일 내의 이미지 파일만 S3에 업로드
				if isImageFile(imageFile.Name) {
					// 파일 확장자 추출
					var fileExtensionName string
					for idx:=0; idx<len(imageFile.Name); idx++ {
						if(imageFile.Name[idx] == '.'){
							fileExtensionName = imageFile.Name[idx:]
							// log.Printf("filename => %v\t fileExtensionName => %v", imageFile.Name, fileExtensionName)
							break
						}
					}
					// S3 업로드 로직
					uploader := s3manager.NewUploader(sess)
					uuid := uuid.New()
					uploadOutput, err := uploader.Upload(&s3manager.UploadInput{
						Bucket: aws.String("kkamji-image-upload-test"),
						Key:    aws.String(fmt.Sprintf("%v%v", uuid.String(), fileExtensionName)),
						Body:   zipFileReader,
					})
					if err != nil {
						log.Fatal("Error upload to S3: ", err)
						continue
					}
					imageURL := uploadOutput.Location // S3에 업로드된 이미지 URL
					log.Printf("S3 image URL => %v", imageURL)
					uploadFileCount++
				} else {
					log.Printf("This file isn't photo\n")
				}
			}
		} else if isImageFile(file.Filename) {
			// 파일 확장자 추출
			var fileExtensionName string
			for idx:=0; idx<len(file.Filename); idx++ {
				// log.Println(len(file.Filename))
				if(file.Filename[idx] == '.'){
					fileExtensionName = file.Filename[idx:]
					// log.Printf("filename => %v\t fileExtensionName => %v\n", file.Filename, fileExtensionName)
					break
				}
			}
			// 파일 전송
			uploader := s3manager.NewUploader(sess)
			uuid := uuid.New()
			uploadOutput, err := uploader.Upload(&s3manager.UploadInput{
				Bucket: aws.String("kkamji-image-upload-test"),
				Key:    aws.String(fmt.Sprintf("%v%v",uuid.String(), fileExtensionName)),
				Body:   src,
			})
			if err != nil {
				c.JSON(500, gin.H{"message": fmt.Sprintf("Unable to upload file: %v", err)})
				return
			}
			imageURL := uploadOutput.Location // S3에 업로드된 이미지 URL
			log.Printf("S3 image URL => %v\n", imageURL)
			uploadFileCount++
		} else {
			log.Printf("This file isn't photo\n")
		}
	}
	// currentTime := time.Now()
	// RDS에 데이터 저장
	// _, err = db.Exec("INSERT INTO posts (user_id, image_url, content, create_at, update_at) VALUES (?, ?, ?, ?, ?)",
	// 	post.UserId, imageURL, post.Content, currentTime, currentTime)
	// if err != nil {
	// 	log.Printf("post.UserId => %v", post.UserId)   // UserId 확인용 코드
	// 	log.Printf("post.Content => %v", post.Content) // Content 확인용 코드
	// 	c.JSON(500, gin.H{"message": fmt.Sprintf("Unable to save post data: %v", err)})
	// 	return
	// }
	log.Printf("%v file uploaded", uploadFileCount)
	if len(fileHeader) == 0 {
		c.JSON(200, gin.H{"message": "no file in"})
	} else {
		c.JSON(200, gin.H{"message": "File uploaded successfully"})
	}
}

func isImageFile(fileName string) bool {
    // 이미지 파일 확장자를 확인하는 간단한 로직
    // 실제 사용시에는 더 많은 이미지 형식을 확인할 수 있도록 확장 필요
    return strings.HasSuffix(fileName, ".png") || strings.HasSuffix(fileName, ".jpg") || strings.HasSuffix(fileName, ".jpeg")
}