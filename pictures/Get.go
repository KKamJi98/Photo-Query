package picture

import (
	"ace-app/databases"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)



// GetPictures retrieves all pictures from the database.
func GetPictures(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	var pictures []Picture

	// Query all pictures from the database.
	rows, err := db.Query("SELECT picture_id, user_id, image_url, create_at, delete_at, bookmarked FROM Pictures")
	if err != nil {
		log.Printf("Error querying pictures: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying pictures"})
		return
	}
	defer rows.Close()

	// Scan each row into the Picture struct.
	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.DeletedAt, &picture.Bookmarked); err != nil {
			log.Printf("Error scanning picture: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	c.JSON(http.StatusOK, gin.H{"pictures": pictures})
}

// GetPicturesByUserId retrieves pictures for a specific user ID.
func GetPicturesByUserId(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	var pictures []Picture
	userId := c.Param("user_id")

	// Query pictures by user ID.
	rows, err := db.Query("SELECT picture_id, user_id, image_url, create_at, delete_at, bookmarked FROM Pictures WHERE user_id = ?", userId)
	if err != nil {
		log.Printf("Error querying pictures for user %v: %v", userId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying pictures"})
		return
	}
	defer rows.Close()

	// Scan each row into the Picture struct.
	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.DeletedAt, &picture.Bookmarked); err != nil {
			log.Printf("Error scanning picture: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	c.JSON(http.StatusOK, gin.H{"pictures": pictures})
}

// GetPictureByPictureId retrieves a single picture by its ID.
func GetPictureByPictureId(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	var picture Picture
	pictureId := c.Param("picture_id")

	// Query a single picture by picture ID.
	err := db.QueryRow("SELECT picture_id, user_id, image_url, create_at, delete_at, bookmarked FROM Pictures WHERE picture_id = ?", pictureId).Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.DeletedAt, &picture.Bookmarked)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "No picture found with given ID"}) // No picture found.
			return
		}
		log.Printf("Error querying picture %v: %v", pictureId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying picture"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"picture": picture})
}
