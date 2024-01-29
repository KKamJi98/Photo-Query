package post

// import (
// 	"database/sql"
// 	"fmt"
// 	"github.com/gin-gonic/gin"
// 	"github.com/joho/godotenv"
// 	"log"
// 	"net/http"
// 	"os"
// )

// func DeletePostByPostId(c *gin.Context) {
// 	db := database.ConnectDB()
// 	defer db.Close()

// 	postId := c.Param("postId")
// 	result, err := db.Exec("DELETE FROM posts WHERE post_id = ?", postId)
// 	if err != nil {
// 		log.Printf("Error deleting post: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting post"})
// 		return
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		log.Printf("Error getting rows affected: %v", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting rows affected"})
// 		return
// 	}

// 	if rowsAffected == 0 {
// 		c.JSON(http.StatusNotFound, gin.H{"message": "No post found with given ID"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
// }
