package picture

import (
	"ace-app/databases"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strconv"
	"math"
)

// GetPictures 함수는 데이터베이스에서 모든 사진을 가져옵니다.
func GetPictures(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	var pictures []Picture

	// 데이터베이스에서 모든 사진을 조회합니다.
	rows, err := db.Query("SELECT picture_id, user_id, image_url, create_at, bookmarked FROM Pictures")
	if err != nil {
		log.Printf("사진 조회 오류: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
		return
	}
	defer rows.Close()

	// 각 행을 Picture 구조체로 스캔합니다.
	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.Bookmarked); err != nil {
			log.Printf("사진 스캔 오류: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	c.JSON(http.StatusOK, gin.H{"pictures": pictures})
}

// GetPicturesByUserId 함수는 특정 사용자 ID에 대한 사진을 가져옵니다.
func GetPicturesByUserId(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	var pictures []Picture
	userId := c.Param("user_id")

	limit := c.Query("limit")
	if limit == "" {
		limit = "20"
	}
	log.Printf("%s", limit)

	last := c.Query("last")
	if last == "" {
		var pictureID string
		// 단일 행 조회에는 QueryRow를 사용, Scan 메소드로 결과 추출
		err := db.QueryRow("SELECT picture_id FROM Pictures WHERE user_id = ? ORDER BY picture_id DESC LIMIT 1", userId).Scan(&pictureID)
		if err != nil {
			log.Printf("Error retrieving last picture: %v", err)
			// return // 에러 시 함수 종료 혹은 적절한 에러 처리
			pictureID = strconv.Itoa(math.MaxInt64 - 1)
		}
		lastIndex, err := strconv.Atoi(pictureID)
		if err != nil {
			log.Printf("Error converting pictureID to int: %v", err)
			return // 변환 에러 시 함수 종료 혹은 적절한 에러 처리
		}
		lastIndex += 1                 // lastIndex 값을 1 증가
		last = strconv.Itoa(lastIndex) // 다시 문자열로 변환
	}
	log.Printf("last: %s", last)

	// 사용자 ID별로 사진을 조회합니다.
	rows, err := db.Query("SELECT picture_id, user_id, image_url, create_at, bookmarked FROM Pictures WHERE (user_id = ? AND picture_id < ?) ORDER BY picture_id DESC LIMIT ?", userId, last, limit)
	if err != nil {
		log.Printf("사용자 %v에 대한 사진 조회 오류: %v", userId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
		return
	}
	defer rows.Close()

	// 각 행을 Picture 구조체로 스캔합니다.
	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.Bookmarked); err != nil {
			log.Printf("사진 스캔 오류: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	c.JSON(http.StatusOK, gin.H{"pictures": pictures})
}

// GetPictureByPictureId 함수는 ID로 단일 사진을 가져옵니다.
func GetPictureByPictureId(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	var pictures []Picture
	pictureId := c.Param("picture_id")

	// 사진 ID로 사진을 조회합니다.
	rows, err := db.Query("SELECT picture_id, user_id, image_url, create_at, bookmarked FROM Pictures WHERE picture_id = ?", pictureId)
	if err != nil {
		log.Printf("사진 조회 오류: %v, %v", pictureId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
		return
	}
	defer rows.Close()

	// 각 행을 Picture 구조체로 스캔합니다.
	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.Bookmarked); err != nil {
			log.Printf("사진 스캔 오류: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	c.JSON(http.StatusOK, gin.H{"pictures": pictures})
}
