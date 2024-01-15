package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"database/sql"
	// "fmt"
	// "log"
	_ "github.com/go-sql-driver/mysql"
)


func main() {
	db, err := sql.Open("mysql", "root:1234@tcp(localhost:3306)/world")

	router := gin.Default()
	
	router.POST("/api-v1/posts") // 포스트 생성
	router.GET("/api-v1/posts") // 포스트 조회
	router.PATCH("/api-v1/posts") // 포스트 수정
	router.DELETE("/api-v1/posts") // 포스트 삭제
	router.POST("/api-v1/posts/{postId}/likes") // 좋아요 기능
	router.GET("/api-v1/posts/{postId}/likes") // 좋아요 조회 (Post 기준)
	router.GET("/api-v1/users/{postId}/likes") // 좋아요 조회 (User 기준)
	router.GET("/api-v1/posts/{postId}/rank") // 포스트 랭킹 조회

	router.Run(":8080")
}