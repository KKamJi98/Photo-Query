package bookmark

import (
	"ace-app/databases"
	"ace-app/pictures"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Bookmark(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	pictureID := c.Param("picture_id")

	var picture picture.Picture
	// 쿼리 실행
	err := db.QueryRow("SELECT bookmarked FROM Pictures WHERE picture_id = ?", pictureID).Scan(&picture.Bookmarked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
		return
	}

	if picture.Bookmarked == 0 {
		picture.Bookmarked = 1
	} else {
		picture.Bookmarked = 0
	}

	result, err := db.Exec("UPDATE Pictures SET bookmarked =? WHERE picture_id =?", picture.Bookmarked, pictureID)
	if err != nil {
		log.Fatal(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "북마크 반영 실패 "})
		return
	}
    log.Printf("affacted %v\t%T", result, result)

    c.JSON(http.StatusOK, gin.H{"message": "북마크 반영 성공", "bookmarked": picture.Bookmarked})
}
