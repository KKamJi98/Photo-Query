package picture

import (
	"database/sql"
	"ace-app/databases"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

type Picture struct {
	PictureID 	int64        `json:"picture_id"`
	UserID    	int64        `json:"user_id"`
	ImageURL  	string       `json:"image_url"`
	CreatedAt 	sql.NullTime `json:"created_at"`
	DeletedAt 	sql.NullTime `json:"deleted_at"`
	Bookmarked 	int8		 `json:"bookmarked`
}

func GetPictures(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()
	
	var pictures []Picture

	rows, err := db.Query("SELECT picture_id, user_id, image_url, create_at, delete_at, bookmarked FROM Pictures")
	if err != nil {
		log.Printf("Error querying pictures: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying pictures"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.DeletedAt, &picture.Bookmarked); err != nil {
			log.Printf("Error scanning picturet: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	c.JSON(http.StatusOK, gin.H{
		"pictures": pictures,
	})
}

func GetPicturesByUserId(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	// 포스트 추출
	var pictures []Picture
	userId := c.Param("user_id")
	rows, err := db.Query("SELECT picture_id, user_id, image_url, create_at, delete_at, bookmarked FROM Pictures WHERE user_id = ?", userId)
	if err != nil {
		log.Printf("Error querying pictures for user %v: %v", userId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying pictures"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.DeletedAt, &picture.Bookmarked); err != nil {
			log.Printf("Error scanning picture: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	c.JSON(http.StatusOK, gin.H{
		"pictures": pictures,
	})
}

func GetPictureByPictureId(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	// 포스트 추출
	var picture Picture
	pictureId := c.Param("picture_id")
	// row, err := db.QueryRow("SELECT picture_id, user_id, image_url, create_at, delete_at, bookmarked FROM Pictures WHERE picture_id = ?", pictureId)
	err := db.QueryRow("SELECT picture_id, user_id, image_url, create_at, delete_at, bookmarked FROM Pictures WHERE picture_id = ?", pictureId).Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.DeletedAt, &picture.Bookmarked)
	if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "No picture found with given ID"})
            return
        }
        log.Printf("Error querying picture %v: %v", pictureId, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying picture"})
        return
    }

	c.JSON(http.StatusOK, gin.H{
		"picture": picture,
	})
}
