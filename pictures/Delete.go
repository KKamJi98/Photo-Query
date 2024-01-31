package picture

import (
	"ace-app/databases"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// DeletePicturesByPostId handles the deletion of pictures by their IDs.
func DeletePicturesByPostId(c *gin.Context) {
	// Define a struct to match the expected JSON data format.
	type picture struct {
		PictureID int64 `json:"picture_id"`
	}

	var pictures []picture

	// Extract and unmarshal the JSON data from the request.
	jsonData := c.PostForm("json_data")
	if err := json.Unmarshal([]byte(jsonData), &pictures); err != nil {
		log.Printf("Invalid JSON data: %v", err)
		c.JSON(400, gin.H{"message": "Invalid JSON data", "error": err.Error()})
		return
	}

	db := database.ConnectDB()
	defer db.Close()

	// Initialize counters for tracking the outcome of deletion operations.
	errorCount := 0
	successCount := 0
	nofoundCount := 0

	// Iterate over each picture ID and attempt to delete it.
	for _, pic := range pictures {
		result, err := db.Exec("DELETE FROM Pictures WHERE picture_id = ?", pic.PictureID)
		if err != nil {
			log.Printf("Error deleting picture with ID %d: %v", pic.PictureID, err)
			errorCount++
			continue
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("Error getting rows affected for picture ID %d: %v", pic.PictureID, err)
			errorCount++
			continue
		}

		if rowsAffected == 0 {
			log.Printf("No picture found with ID %d", pic.PictureID)
			nofoundCount++
		} else {
			successCount++
		}
	}

	// Log the outcome of the operation.
	log.Printf("%d pictures deleted || %d error occurred || %d no picture ID found", successCount, errorCount, nofoundCount)

	// Respond with the outcome of the deletion operation.
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d pictures deleted, %d errors occurred, %d picture IDs not found", successCount, errorCount, nofoundCount)})
}
