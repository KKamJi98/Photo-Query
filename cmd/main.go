package main

import (
	"ace-app/pictures" // picture 패키지 임포트
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load("./env/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_ENDPOINT"), os.Getenv("DB_NAME"), os.Getenv("BUCKET_NAME"))
	cwd, _ := os.Getwd()
	log.Println("Current working directory:", cwd)

	r := gin.Default()

	// PostCreate API
	r.POST("/pictures", picture.CreatePictures)
	r.GET("/pictures", picture.GetPictures)
	r.GET("/users/:user_id/pictures", picture.GetPicturesByUserId)
	r.GET("/picture/:picture_id", picture.GetPictureByPictureId)
	r.DELETE("/pictures", picture.DeletePicturesByPostId)

	r.Run(":8080") // 서버 시작
}
