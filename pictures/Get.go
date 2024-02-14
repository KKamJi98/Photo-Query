package picture

import (
	"ace-app/databases"
	"database/sql"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
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
		last = strconv.Itoa(math.MaxInt64)
		log.Printf("%s", "last index not found")
	}
	log.Printf("last: %s", last)

	bookmark := c.Query("bookmark")
	if bookmark == "" {
		bookmark = "0"
	} else {
		bookmark = "1"
	}

	var rows *sql.Rows
	var err error
	// 사용자 ID별로 사진을 조회합니다.
	if bookmark == "0"{
		rows, err = db.Query("SELECT picture_id, user_id, image_url, create_at, bookmarked FROM Pictures WHERE (user_id = ? AND picture_id < ?) ORDER BY picture_id DESC LIMIT ?", userId, last, limit)
		if err != nil {
			log.Printf("사용자 %v에 대한 사진 조회 오류: %v", userId, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
			return
		}
	} else {
		rows, err = db.Query("SELECT picture_id, user_id, image_url, create_at, bookmarked FROM Pictures WHERE (user_id = ? AND picture_id < ? AND bookmarked = 1) ORDER BY picture_id DESC LIMIT ?", userId, last, limit)
		if err != nil {
			log.Printf("사용자 %v에 대한 사진 조회 오류: %v", userId, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
			return
		}
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

// func GetPicuresByBookmarked(c *gin.Context) {
// 	db := database.ConnectDB()
// 	defer db.Close()

// 	limit := c.Query("limit")
// 	if limit == "" {
// 		limit = "20"
// 	}
// 	log.Printf("%s", limit)

// 	last := c.Query("last")
// 	if last == "" {
// 		last = strconv.Itoa(math.MaxInt64)
// 		log.Printf("%s", "last index not found")
// 	}
// 	log.Printf("last: %s", last)
	
// 	var pictures []Picture
// 	userId := c.Param("user_id")

// 	// 사진 ID로 사진을 조회합니다.
// 	rows, err := db.Query("SELECT picture_id, user_id, image_url, create_at, bookmarked FROM Pictures WHERE (user_id = ? AND picture_id < ? AND bookmarked = 1) ORDER BY picture_id DESC LIMIT ?", userId, last, limit)
// 	if err != nil {
// 		log.Printf("사진 조회 오류: %v, %v", userId, err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
// 		return
// 	}
// 	defer rows.Close()

// 	// 각 행을 Picture 구조체로 스캔합니다.
// 	for rows.Next() {
// 		var picture Picture
// 		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.Bookmarked); err != nil {
// 			log.Printf("사진 스캔 오류: %v", err)
// 			continue
// 		}
// 		pictures = append(pictures, picture)
// 	}

// 	c.JSON(http.StatusOK, gin.H{"pictures": pictures})
// }