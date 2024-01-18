package main

import (
	"database/sql"
	"log"
	// "net/http"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"

	"fmt"
	// "log"
	_ "github.com/go-sql-driver/mysql"
	"post"
)

func main() {
	post.PostCreate()
}


// type Post struct {
// 	PostID int64 `json:"post_id"`
// }

// func main() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}

// 	dbUser := os.Getenv("DB_USER")
// 	dbPassword := os.Getenv("DB_PASSWORD")
// 	dbEndpoint := os.Getenv("DB_ENDPOINT")
// 	dbName := os.Getenv("DB_NAME")

// 	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", dbUser, dbPassword, dbEndpoint, dbName))

// 	if db == nil {
// 		fmt.Println("Can't connect database")
// 	}
// 	router := gin.Default()

// 	router.POST("/api-v1/posts")                // 포스트 생성
// 	router.GET("/api-v1/posts")                 // 포스트 조회
// 	router.PATCH("/api-v1/posts")               // 포스트 수정
// 	router.DELETE("/api-v1/posts")              // 포스트 삭제
// 	router.POST("/api-v1/posts/{postId}/likes") // 좋아요 기능
// 	router.GET("/api-v1/posts/{postId}/likes")  // 좋아요 조회 (Post 기준)
// 	router.GET("/api-v1/users/{postId}/likes")  // 좋아요 조회 (User 기준)
// 	router.GET("/api-v1/posts/{postId}/rank")   // 포스트 랭킹 조회

// 	router.Run(":8080")
// }
