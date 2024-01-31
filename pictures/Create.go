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
    "os"
    "strings"
    "sync"
    "time"
	"context"
)

var uploadFileCount int

// CreatePictures handles the uploaded image files and uploads them to AWS S3.
// CreatePictures handles the uploaded image files and uploads them to AWS S3.
func CreatePictures(c *gin.Context) {
    uploadFileCount = 0
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
        c.JSON(500, gin.H{"message": "AWS session error", "error": err.Error()})
        return
    }
    log.Println("AWS session created successfully\t", sess.Config.Credentials)

    var wg sync.WaitGroup
    errChan := make(chan error, len(fileHeader))

    // Plans routines for file processing.
    var filesPerRoutine int
    if len(fileHeader) < 8 {
        filesPerRoutine = len(fileHeader)
    } else {
        filesPerRoutine = (len(fileHeader)+7) / 8
    }
    log.Printf("Processing files in batches of %d", filesPerRoutine)

    // Performs parallel processing for each file.
    for i := 0; i < len(fileHeader); i += filesPerRoutine {
        end := i + filesPerRoutine
        if end > len(fileHeader) {
            end = len(fileHeader)
        }

        wg.Add(1)
        log.Printf("wg1 called => Processing files %d to %d", i, end-1)

        go func(files []*multipart.FileHeader) {
            defer wg.Done()
            for _, file := range files {
                processFile(file, sess, errChan, picture)
            }
        }(fileHeader[i:end])
    }

    wg.Wait()
    close(errChan)

    // Checks the error channel and logs errors.
    for err := range errChan {
        if err != nil {
            log.Printf("Error in file processing: %v", err)
        }
    }

    log.Printf("%d files uploaded successfully", uploadFileCount)
    c.JSON(200, gin.H{"message": fmt.Sprintf("%v files processing completed", uploadFileCount)})
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
        if numOfFiles < 8 {
            numOfFiles = len(zipReader.File)
        } else {
            numOfFiles = (len(zipReader.File)+7) / 8
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
	log.Printf("uuid=> %v session => %v", uuid.String(), sess.Config.Credentials)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    fileExtension := getFileExtension(fileName)

    uploadOutput, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
        Bucket: aws.String(s3BucketName),
        Key:    aws.String(fmt.Sprintf("%v%v%v", "original/", uuid.String(), fileExtension)),
        Body:   fileReader,
    })

	if err != nil {
        // 타임아웃 에러 확인
        if err == context.DeadlineExceeded {
            log.Printf("Upload timed out: %v", err)
        } else {
            log.Printf("Error in upload: %v", err)
        }
        errChan <- err
        return
    }
	
	log.Printf("uploadOutput=> %v", uploadOutput)
    if err != nil {
        errChan <- err
        return
    }

    // Connects to the database and saves image information.
    db := database.ConnectDB()
    defer db.Close()
    currentTime := time.Now()
    imageURL := uploadOutput.Location

    _, err2 := db.Exec("INSERT INTO Pictures (user_id, image_url, create_at, bookmarked) VALUES (?, ?, ?, ?)",
        pic.UserID, imageURL, currentTime, 0)
    if err2 != nil {
        errChan <- err2
        return
    }

    log.Printf("file upload Complete")
    uploadFileCount++
}

// isImageFile checks if the file name indicates an image file.
func isImageFile(fileName string) bool {
    return strings.HasSuffix(fileName, ".png") || strings.HasSuffix(fileName, ".jpg") || strings.HasSuffix(fileName, ".jpeg")
}

// getFileExtension extracts the file extension from the file name.
func getFileExtension(fileName string) string {
    for i := range fileName {
        if fileName[i] == '.' {
            return fileName[i:]
        }
    }
    return ""
}
