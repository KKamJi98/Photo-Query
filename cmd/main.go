package main

import (
	"ace-app/post" // post 패키지 임포트
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// PostCreate API
	r.POST("/api/v1/posts", post.CreatePost)
	r.GET("/api/v1/posts", post.GetPosts)
	r.GET("/api/v1/posts/:postId", post.GetPostByPostId)
	r.GET("/api/v1/users/:userId/posts", post.GetPostsByUserId)
	r.DELETE("/api/v1/posts/:postId", post.DeletePostByPostId)

	r.Run(":8080") // 서버 시작 
}
