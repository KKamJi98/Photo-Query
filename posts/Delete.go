package post

import (
	"ace-app/databases"
	// "database/sql"
	"encoding/json"
	// "fmt"
	"github.com/gin-gonic/gin"
	// "github.com/joho/godotenv"
	"log"
	"net/http"
	"fmt"
	// "os"
)

func DeletePicturesByPostId(c *gin.Context) {
	type picture struct {
		PictureId  int64  `json:"picture_id"`
	}
	
	var pictures []picture
	
	jsonData := c.PostForm("json_data")
	if err := json.Unmarshal([]byte(jsonData), &pictures); err != nil {
		c.JSON(400, gin.H{"message": "Invalid JSON data", "error": err.Error()})
		return
	}

	db := database.ConnectDB()
	defer db.Close()

	// postId := c.Param("postId")

	errorCount := 0
	successCount := 0
	nofoundCount := 0
	for _, pic := range pictures {
		result, err := db.Exec("DELETE FROM Pictures WHERE picture_id = ?", pic.PictureId)
		if err != nil {
			log.Printf("Error deleting picture: %v", err)
		}
	
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("Error getting rows affected: %v", err)
			log.Printf("picture_id => %v", pic.PictureId)
			errorCount++
			// c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting rows affected"})
		} 
	
		if rowsAffected == 0 {
			log.Printf("No picture found with given ID")
			log.Printf("picture_id => %v", pic.PictureId)
			nofoundCount++
			// c.JSON(http.StatusNotFound, gin.H{"message": "No picture found with given ID"})
			// return
		} else {
			successCount++
		}
	}
	log.Printf("%d pictures deleted || %d error occurred || %d no post_id selected", successCount, errorCount, nofoundCount)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d pictures deleted && %d error occurred %d no post_id selected", successCount, errorCount, nofoundCount)})//"picture deleted successfully"})
}
