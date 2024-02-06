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

// CreatePictures는 업로드된 이미지 파일을 처리하고 AWS S3에 업로드합니다.
func CreatePictures(c *gin.Context) {
	uploadFileCount = 0
	urls = []string{}

	var picture Picture
	jsonData := c.PostForm("json_data")

	// JSON 데이터를 Picture 구조체로 언마샬합니다.
	if err := json.Unmarshal([]byte(jsonData), &picture); err != nil {
		log.Printf("JSON 데이터 언마샬 오류: %v", err)
		c.JSON(400, gin.H{"message": "잘못된 JSON 데이터", "error": err.Error()})
		return
	}
	log.Println("JSON 데이터 언마샬 성공")

	// 멀티파트 폼 데이터를 수신합니다.
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("멀티파트 폼 데이터 수신 오류: %v", err)
		c.JSON(500, gin.H{"message": "파일 수신 오류"})
		return
	}
	log.Println("멀티파트 폼 데이터 수신 완료")
	fileHeader := form.File["images"]

	// AWS 세션을 생성합니다.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		log.Printf("AWS 세션 생성 오류: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": 1000, "message": "aws 세션을 찾을 수 없음"}) // 1000번 에러 코드 반환
		return
	}
	log.Println("AWS 세션 생성 성공\t", sess.Config.Credentials)
	
	var wg sync.WaitGroup
	errChan := make(chan error, len(fileHeader))
	
	var filesPerRoutine int
	
	if len(fileHeader) < 32 {
		filesPerRoutine = len(fileHeader)
	} else {
		filesPerRoutine = (len(fileHeader) + 31) / 32
	}
	log.Printf("파일 일괄 처리 크기: %d", filesPerRoutine)
		
	// 각 파일에 대한 병렬 처리를 수행합니다.
	for i := 0; i < len(fileHeader); i += filesPerRoutine {
		end := i + filesPerRoutine
		if end > len(fileHeader) {
			end = len(fileHeader)
		}

		wg.Add(1)
		log.Printf("wg 호출\t 파일 처리 중 %d부터 %d까지", i, end-1)

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
		log.Println("wg1 => 모든 파일 처리 루틴 완료")
	}()

	for err := range errChan {
		if err != nil {
			log.Printf("파일 처리 오류: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": 2000, "message": fmt.Sprintf("%v", err)}) // 2000번 에러 코드 반환
			return
		}
	}

	log.Printf("%d개 파일 업로드 성공", uploadFileCount)
	if uploadFileCount == 0 {
		c.JSON(200, gin.H{"message": fmt.Sprintf("%v개 파일 처리 완료", uploadFileCount)})
	} else {
		GetPicturesByUrls(c, urls)
	}
}

// processFile는 개별 파일 처리 및 S3에 업로드합니다.
func processFile(file *multipart.FileHeader, sess *session.Session, errChan chan<- error, pic Picture) {

	src, err := file.Open()
	if err != nil {
		errChan <- err
		return
	}
	if src == nil {
		errChan <- errors.New("파일 리더가 nil입니다")
		return
	}
	defer src.Close()

	// ZIP 파일 처리
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
			log.Printf("wg2 호출\t 파일 처리 중 %d부터 %d까지", i, end-1)
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
					log.Println("wg2 => ZIP 파일 처리 루틴의 마지막 배치 완료") // wg2가 마지막 배치에서 완료된 후 로그
				}
			}(zipReader.File[i:end])
		}
		wg2.Wait()
	} else if isImageFile(file.Filename) {
		uploadToS3(src, file.Filename, sess, errChan, pic)
	}
}

// uploadToS3 함수는 파일을 AWS S3에 업로드합니다.
func uploadToS3(fileReader io.Reader, fileName string, sess *session.Session, errChan chan<- error, pic Picture) {
	if fileReader == nil {
		errChan <- errors.New("fileReader가 nil입니다")
		return
	}
	s3BucketName := os.Getenv("BUCKET_NAME")

	uploader := s3manager.NewUploader(sess)
	uuid := uuid.New()

	fileExtension := getFileExtension(fileName)

	// uploadOutput, err := uploader.Upload(&s3manager.UploadInput{
	// 	Bucket: aws.String(s3BucketName),
	// 	Key:    aws.String(fmt.Sprintf("%v%v%v", "original/", uuid.String(), fileExtension)),
	// 	Body:   fileReader,
	// })

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(fmt.Sprintf("%v%v%v", "original/", uuid.String(), fileExtension)),
		Body:   fileReader,
	})

	if err != nil {
		log.Printf("업로드 중 오류 발생: %v", err)
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
	imageURL := uuid.String() + fileExtension
	// imageURL := uploadOutput.Location
	urls = append(urls, imageURL)

	_, err2 := db.Exec("INSERT INTO Pictures (user_id, image_url, create_at, bookmarked) VALUES (?, ?, ?, ?)",
		pic.UserID, imageURL, currentTime, 0)
	if err2 != nil {
		errChan <- err2
		return
	}

	uploadFileCount++
	log.Printf("%v 파일 업로드 완료", uploadFileCount)
}

// isImageFile 함수는 파일 이름이 이미지 파일을 나타내는지 확인합니다.
func isImageFile(fileName string) bool {
	// 지원하는 이미지 파일 확장자를 추가합니다.
	validExtensions := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp"}
	for _, ext := range validExtensions {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}
	return false
}

// getFileExtension 함수는 파일 이름에서 확장자를 추출합니다.
func getFileExtension(fileName string) string {
	// 파일 이름에서 마지막으로 나타나는 '.'의 위치를 찾아 확장자를 반환합니다.
	if dotIndex := strings.LastIndex(fileName, "."); dotIndex != -1 {
		return fileName[dotIndex:]
	}
	return "" // 확장자가 없는 경우
}

// GetPicturesByUrls 함수는 URL을 기반으로 Picture 정보를 검색하고 반환합니다.
func GetPicturesByUrls(c *gin.Context, urls []string) {
	db := database.ConnectDB()
	defer db.Close()

	var pictures []Picture

	// 동적으로 IN 절 쿼리를 생성합니다.
	placeholders := make([]string, len(urls))
	args := make([]interface{}, len(urls))
	for i, url := range urls {
		placeholders[i] = "?"
		args[i] = url
	}
	query := fmt.Sprintf("SELECT picture_id, user_id, image_url, create_at, bookmarked FROM Pictures WHERE image_url IN (%s)", strings.Join(placeholders, ","))

	// 쿼리를 실행합니다.
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("URL을 기반으로 사진을 쿼리하는 중 오류 발생: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 쿼리 오류"})
		return
	}
	defer rows.Close()

	// 결과를 스캔합니다.
	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.Bookmarked); err != nil {
			log.Printf("사진 스캔 중 오류 발생: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	// 클라이언트에게 JSON 형식으로 결과를 반환합니다.
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%v 파일 업로드 성공", uploadFileCount) ,"pictures": pictures})
}