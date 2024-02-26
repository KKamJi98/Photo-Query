package picture

import (
	"ace-app/databases"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// DeletePicturesByPostId는 그림들을 ID로 삭제하는 작업을 처리합니다.
func DeleteAllPictures(c *gin.Context) {

	// 예상되는 JSON 데이터 형식과 일치하는 구조체를 정의합니다.
	// type picture struct {
	// 	PictureID int64 `json:"picture_id"`
	// }

	
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
	
	// 각 그림 ID를 반복하고 삭제를 시도합니다.
	result, err := db.Exec("DELETE FROM Pictures WHERE user_id = ?", userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message" : "사진 삭제 실패"})
	}
	rowsAffected, err := result.RowsAffected()
	s3DeleteAllPictures(pictures, db)
	log.Printf("%d개 사진 삭제 완료", rowsAffected)
	// 삭제 작업 결과를 응답합니다.
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d개 사진 삭제 완료", rowsAffected)})
}

func s3DeleteAllPictures(pictures []Picture, db *sql.DB) {
	// db := database.ConnectDB()
	// defer db.Close()
	// TODO
	//* picutre_id를 사용해서 이미지 이름, user_id 찾기
	//* 이것을 키값으로 S3 Bucket에서 이미지 삭제
	// var imageName, userId string
	var s3ImageOriginalObjectKeys []string
	var s3ImageThumbnailObjectKeys []string

	log.Println(len(pictures))
	for _, pic := range pictures {
		// query := db.QueryRow("SELECT image_url, user_id FROM Pictures WHERE picture_id =?", pic.PictureID)
		// err := query.Scan(&imageName, &userId)
		// if err != nil {
		// 	if err == sql.ErrNoRows {
		// 		log.Printf("No results found.")
		// 	} else {
		// 		log.Printf("Query failed: %v", err)
		// 	}
		// }
		log.Printf("%v/%v", pic.ImageURL, pic.PictureID)
		s3ImageOriginalObjectKeys = append(s3ImageOriginalObjectKeys, fmt.Sprintf("%s/%s/%s", "original", pic.UserID, pic.ImageURL))
		s3ImageThumbnailObjectKeys = append(s3ImageThumbnailObjectKeys, fmt.Sprintf("%s/%s/%s", "thumbnail", pic.UserID, pic.ImageURL))
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
	basics.DeleteObjects("rapa-app-image-bucket", s3ImageOriginalObjectKeys)
	basics.DeleteObjects("rapa-app-image-bucket", s3ImageThumbnailObjectKeys)
}


