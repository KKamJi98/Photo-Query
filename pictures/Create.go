package picture

import (
	"ace-app/databases"
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var uploadFileCount int
var urls []string

// CreatePictures handles the uploaded image files and uploads them to AWS S3.
func CreatePictures(c *gin.Context) {
	uploadFileCount = 0
	urls = []string{}

	var picture Picture
	jsonData := c.PostForm("json_data")

	// Unmarshals the JSON data into the Picture struct.
	if err := json.Unmarshal([]byte(jsonData), &picture); err != nil {
		log.Printf("Error unmarshaling JSON data: %v", err)
		c.JSON(400, gin.H{"message": "Invalid JSON data", "error": err.Error()})
		return
	}
	log.Println("JSON data unmarshaled successfully")

	// Receives multipart form data.
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Error receiving multipart form data: %v", err)
		c.JSON(500, gin.H{"message": "File receive error"})
		return
	}
	log.Println("Multipart form data received")
	fileHeader := form.File["images"]

	// Creates an AWS session.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		log.Printf("Error creating AWS session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": 1000, "message": "aws session can not found"}) // 1000번 에러 코드 반환
		return
	}
	log.Println("AWS session created successfully\t", sess.Config.Credentials)
	
	var wg sync.WaitGroup
	errChan := make(chan error, len(fileHeader))
	
	var filesPerRoutine int
	
	if len(fileHeader) < 32 {
		filesPerRoutine = len(fileHeader)
		} else {
			filesPerRoutine = (len(fileHeader) + 31) / 32
		}
		log.Printf("Processing files in batches of %d", filesPerRoutine)
		
	// Performs parallel processing for each file.
	for i := 0; i < len(fileHeader); i += filesPerRoutine {
		end := i + filesPerRoutine
		if end > len(fileHeader) {
			end = len(fileHeader)
		}

		wg.Add(1)
		log.Printf("wg Called Processing files %d to %d", i, end-1)

		go func(files []*multipart.FileHeader) {
			defer wg.Done()
			for _, file := range files {
				processFile(file, sess, errChan, picture)
			}
		}(fileHeader[i:end])
	}

	// 고루틴 대기 및 에러 채널 처리
	go func() {
		wg.Wait()
		close(errChan)
		log.Println("wg1 => All file processing routines have completed")
	}()

	for err := range errChan {
		if err != nil {
			log.Printf("Error in file processing: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": 2000, "message": "Go Routine Error"}) // 2000번 에러 코드 반환
			return
		}
	}

	log.Printf("%d files uploaded successfully", uploadFileCount)
	if uploadFileCount == 0 {
		c.JSON(200, gin.H{"message": fmt.Sprintf("%v files processing completed", uploadFileCount)})
	} else {
		GetPicturesByUrls(c, urls)
	}
}

// processFile handles individual file processing and uploads to S3.
func processFile(file *multipart.FileHeader, sess *session.Session, errChan chan<- error, pic Picture) {

	src, err := file.Open()
	if err != nil {
		errChan <- err
		return
	}
	if src == nil {
		errChan <- errors.New("file reader is nil")
		return
	}
	defer src.Close()

	// Handles ZIP file processing.
	if strings.HasSuffix(file.Filename, ".zip") {
		zipReader, err := zip.NewReader(src, file.Size)
		if err != nil {
			errChan <- err
			return
		}
		numOfFiles := len(zipReader.File)
		if numOfFiles < 32 {
			numOfFiles = len(zipReader.File)
		} else {
			numOfFiles = (len(zipReader.File) + 31) / 32
		}
		var wg2 sync.WaitGroup
		for i := 0; i < len(zipReader.File); i += numOfFiles {
			end := i + numOfFiles
			if end > len(zipReader.File) {
				end = len(zipReader.File)
			}
			wg2.Add(1)
			log.Printf("wg2 called\t Processing files %d to %d", i, end-1)
			go func(files []*zip.File) {
				defer wg2.Done()
				for _, file := range files {
					if isImageFile(file.Name) {
						zipFileReader, err := file.Open()
						if err != nil {
							errChan <- err
							continue
						}
						defer zipFileReader.Close()
						uploadToS3(zipFileReader, file.Name, sess, errChan, pic)
					}
				}
				if end == len(zipReader.File) {
					log.Println("wg2 => Final batch of ZIP file processing routines have completed") // wg2가 마지막 배치에서 완료된 후 로그
				}
			}(zipReader.File[i:end])
		}
		wg2.Wait()
	} else if isImageFile(file.Filename) {
		uploadToS3(src, file.Filename, sess, errChan, pic)
	}
}

// uploadToS3 uploads a file to AWS S3.
func uploadToS3(fileReader io.Reader, fileName string, sess *session.Session, errChan chan<- error, pic Picture) {
	if fileReader == nil {
		errChan <- errors.New("fileReader is nil")
		return
	}
	s3BucketName := os.Getenv("BUCKET_NAME")

	uploader := s3manager.NewUploader(sess)
	uuid := uuid.New()

	fileExtension := getFileExtension(fileName)

	uploadOutput, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(fmt.Sprintf("%v%v%v", "original/", uuid.String(), fileExtension)),
		Body:   fileReader,
	})

	if err != nil {
		log.Printf("Error in upload: %v", err)
		errChan <- err
		return
	}

	if err != nil {
		errChan <- err
		return
	}

	db := database.ConnectDB()
	defer db.Close()
	currentTime := time.Now()
	imageURL := uploadOutput.Location
	urls = append(urls, imageURL)

	_, err2 := db.Exec("INSERT INTO Pictures (user_id, image_url, create_at, bookmarked) VALUES (?, ?, ?, ?)",
		pic.UserID, imageURL, currentTime, 0)
	if err2 != nil {
		errChan <- err2
		return
	}

	uploadFileCount++
	log.Printf("%v file upload Complete", uploadFileCount)
}

// isImageFile checks if the file name indicates an image file.
func isImageFile(fileName string) bool {
	// 지원하는 이미지 파일 확장자 추가
	validExtensions := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp"}
	for _, ext := range validExtensions {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}
	return false
}

// getFileExtension extracts the file extension from the file name.
func getFileExtension(fileName string) string {
	// 마지막으로 나타나는 '.'의 위치를 찾아 확장자 반환
	if dotIndex := strings.LastIndex(fileName, "."); dotIndex != -1 {
		return fileName[dotIndex:]
	}
	return "" // 확장자가 없는 경우
}

// 파일 업로드 후 URL을 저장한 슬라이스를 기반으로 Picture 정보를 검색하고 반환하는 함수
func GetPicturesByUrls(c *gin.Context, urls []string) {
	db := database.ConnectDB()
	defer db.Close()

	var pictures []Picture

	// 동적으로 IN 절 쿼리 생성
	placeholders := make([]string, len(urls))
	args := make([]interface{}, len(urls))
	for i, url := range urls {
		placeholders[i] = "?"
		args[i] = url
	}
	query := fmt.Sprintf("SELECT picture_id, user_id, image_url, create_at, delete_at, bookmarked FROM Pictures WHERE image_url IN (%s)", strings.Join(placeholders, ","))

	// 쿼리 실행
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying pictures by URLs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying pictures"})
		return
	}
	defer rows.Close()

	// 결과 스캔
	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.DeletedAt, &picture.Bookmarked); err != nil {
			log.Printf("Error scanning picture: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	// 클라이언트에게 JSON 형식으로 결과 반환
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d files uploaded successfully", uploadFileCount) ,"pictures": pictures})
}