package post

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func DeletePostByPostId(c *gin.Context) {
	// 환경변수 불러오기
	err := godotenv.Load("./env/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 환경변수 변수에 저장하기
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbEndpoint := os.Getenv("DB_ENDPOINT")
	dbName := os.Getenv("DB_NAME")

	// 데이터베이스 연결
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", dbUser, dbPassword, dbEndpoint, dbName))
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	defer db.Close()

	// 데이터베이스 연결 테스트
	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to database: ", err)
	}

	postId := c.Param("postId")
	result, err := db.Exec("DELETE FROM posts WHERE post_id = ?", postId)
	if err != nil {
		log.Printf("Error deleting post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting post"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting rows affected"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No post found with given ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
