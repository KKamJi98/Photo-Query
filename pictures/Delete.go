package picture

import (
	"ace-app/databases"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"fmt"
)

func DeletePicturesByPostId(c *gin.Context) {
	type picture struct {
		PictureID  int64  `json:"picture_id"`
	}
	
	var pictures []picture
	
	jsonData := c.PostForm("json_data")
	if err := json.Unmarshal([]byte(jsonData), &pictures); err != nil {
		c.JSON(400, gin.H{"message": "Invalid JSON data", "error": err.Error()})
		return
	}

	db := database.ConnectDB()
	defer db.Close()

	errorCount := 0
	successCount := 0
	nofoundCount := 0
	for _, pic := range pictures {
		result, err := db.Exec("DELETE FROM Pictures WHERE picture_id = ?", pic.PictureID)
		if err != nil {
			log.Printf("Error deleting picture: %v", err)
		}
	
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("Error getting rows affected: %v", err)
			log.Printf("picture_id => %v", pic.PictureID)
			errorCount++
			// c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting rows affected"})
		} 
	
		if rowsAffected == 0 {
			log.Printf("No picture found with given ID")
			log.Printf("picture_id => %v", pic.PictureID)
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
