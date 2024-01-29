package main

import (
	"ace-app/pictures" // picture 패키지 임포트
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// PostCreate API
	r.POST("/pictures", picture.CreatePictures)
	r.GET("/pictures", picture.GetPictures)
	r.GET("/users/:user_id/pictures", picture.GetPicturesByUserId)
	r.GET("/picture/:picture_id", picture.GetPictureByPictureId)
	r.DELETE("/pictures", picture.DeletePicturesByPostId)

	r.Run(":8080") // 서버 시작
}
