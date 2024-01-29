package post

import (
	"ace-app/databases"
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strings"
	"sync"
	"time"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var uploadFileCount int
var picture struct {
	UserId  int64  `json:"user_id"`
}

func CreatePost(c *gin.Context) {
	jsonData := c.PostForm("json_data")
	if err := json.Unmarshal([]byte(jsonData), &picture); err != nil {
		c.JSON(400, gin.H{"message": "Invalid JSON data", "error": err.Error()})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(500, gin.H{"message": "File receive error"})
		return
	}
	fileHeader := form.File["images"]

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-2"),
	})
	if err != nil {
		c.JSON(500, gin.H{"message": "AWS session error"})
		return
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(fileHeader))

	var filesPerRoutine int
	if len(fileHeader) < 8 {
		filesPerRoutine = len(fileHeader)
	} else {
		filesPerRoutine = len(fileHeader) / 8
	}

	for i := 0; i < len(fileHeader); i += filesPerRoutine {
		end := i + filesPerRoutine
		if end > len(fileHeader) {
			end = len(fileHeader)
		}

		wg.Add(1)
		log.Println("wg routine called")

		go func(files []*multipart.FileHeader) {
			defer wg.Done()
			for _, file := range files {
				processFile(file, sess, errChan)
			}
		}(fileHeader[i:end])
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}

	log.Println(uploadFileCount, " files upload Complete")
	uploadFileCount = 0
	c.JSON(200, gin.H{"message": "File processing completed"})
}

func processFile(file *multipart.FileHeader, sess *session.Session, errChan chan<- error) {
	src, err := file.Open()
	if err != nil {
		errChan <- err
		return
	}
	defer src.Close()

	if strings.HasSuffix(file.Filename, ".zip") {
		zipReader, err := zip.NewReader(src, file.Size)
		if err != nil {
			errChan <- err
			return
		}
		// log.Println("파일 크기 =>z" ,len(zipReader.File))
		numOfFiles := len(zipReader.File)
		if numOfFiles < 8 {
			numOfFiles = 1
		} else {
			numOfFiles = len(zipReader.File) / 8
		}
		var wg2 sync.WaitGroup
		for i := 0; i < len(zipReader.File); i += numOfFiles {
			end := i + numOfFiles
			if end > len(zipReader.File) {
				end = len(zipReader.File)
			}
			wg2.Add(1)
			go func(files []*zip.File) {
				log.Println("wg2 routine called")
				defer wg2.Done()
				for _, file := range files {
					if isImageFile(file.Name) {
						zipFileReader, err := file.Open()
						if err != nil {
							errChan <- err
							continue
						}
						defer zipFileReader.Close()
						uploadToS3(zipFileReader, file.Name, sess, errChan)
					}
				}
			}(zipReader.File[i:end])
		}
		wg2.Wait()
	} else if isImageFile(file.Filename) {
		uploadToS3(src, file.Filename, sess, errChan)
	}
}

func uploadToS3(fileReader io.Reader, fileName string, sess *session.Session, errChan chan<- error) {
	uploader := s3manager.NewUploader(sess)
	uuid := uuid.New()
	fileExtension := getFileExtension(fileName)
	uploadOutput, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("kkamji-image-upload-test"),
		Key:    aws.String(fmt.Sprintf("%v%v", uuid.String(), fileExtension)),
		Body:   fileReader,
	})
	if err != nil {
		errChan <- err
	}

	db := database.ConnectDB()
	defer db.Close()
	currentTime := time.Now()
	imageURL := uploadOutput.Location

	_, err2 := db.Exec("INSERT INTO Pictures (user_id, image_url, create_at, bookmark) VALUES (?, ?, ?, ?)",
	picture.UserId, imageURL, currentTime, 0)
	if err2 != nil {
		log.Printf("picture.UserId => %v", picture.UserId)   // userId 확인용 코드
		// log.Printf("post.Content => %v", post.Content) // Content 확인용 코드
		// c.JSON(500, gin.H{"message": fmt.Sprintf("Unable to save post data: %v", err)})
		log.Println("Unable to save post data:", err2)
		return
	}

	log.Printf("file upload Complete")
	uploadFileCount++
}

func isImageFile(fileName string) bool {
	return strings.HasSuffix(fileName, ".png") || strings.HasSuffix(fileName, ".jpg") || strings.HasSuffix(fileName, ".jpeg")
}

func getFileExtension(fileName string) string {
	for i := range fileName {
		if fileName[i] == '.' {
			return fileName[i:]
		}
	}
	return ""
}