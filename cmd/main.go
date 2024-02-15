package main

import (
	"ace-app/bookmark"
	"ace-app/pictures"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	// .env 파일에서 환경 변수 로드
	err := godotenv.Load("./env/.env")
	if err != nil {
		log.Fatal("Error loading .env file") // .env 파일 로드 실패 => 로그 출력 && 종료
	}
	// 환경 변수를 로그 출력
	log.Println(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_ENDPOINT"), os.Getenv("DB_NAME"), os.Getenv("BUCKET_NAME"))

	// gin 라우터 인스턴스 생성
	r := gin.Default()

	config := cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"*"},
	}
	r.Use(cors.New(config))

	// RESTful API endpoint 생성
	r.POST("/pictures", picture.CreatePictures)                    // 사진 생성 API
	r.GET("/pictures", picture.GetPictures)                        // 모든 사진 조회 API
	r.GET("/users/:user_id/pictures", picture.GetPicturesByUserId) // 특정 사용자의 사진 조회 API
	r.GET("/picture/:picture_id", picture.GetPictureByPictureId)   // 특정 사진 조회 API
	r.DELETE("/pictures", picture.DeletePicturesByPostId)          // 사진 삭제 API

	r.POST("/pictures/bookmark/:picture_id", bookmark.Bookmark) // 북마크 API
	// r.GET("/users/:user_id/pictures/bookmark", picture.GetPicuresByBookmarked) // 북마크 된 사진 조회 API

	// 8080 포트에서 서버 시작
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err) // 서버 실행 실패 시 로그 출력 및 애플리케이션 종료
	}
}
