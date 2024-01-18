package post

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	// "time"
)

type Post struct {
	PostID    int64        `json:"post_id"`
	UserID    int64        `json:"user_id"`
	ImageURL  string       `json:"image_url"`
	Content   string       `json:"content"`
	CreatedAt sql.NullTime `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

func GetPosts(c *gin.Context) {
	err := godotenv.Load("./env/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbEndpoint := os.Getenv("DB_ENDPOINT")
	dbName := os.Getenv("DB_NAME")

	// 데이터베이스 연결
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", dbUser, dbPassword, dbEndpoint, dbName))
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	defer db.Close()

	// 데이터베이스 연결 테스트
	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to database: ", err)
	}

	var posts []Post

	rows, err := db.Query("SELECT post_id, user_id, image_url, content, create_at, update_at, delete_at FROM posts")
	if err != nil {
		log.Printf("Error querying posts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying posts"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.PostID, &post.UserID, &post.ImageURL, &post.Content, &post.CreatedAt, &post.UpdatedAt, &post.DeletedAt); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}
		posts = append(posts, post)
	}

	// fmt.Println(posts)
	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

func GetPostsByUserId(c *gin.Context) {
	err := godotenv.Load("./env/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbEndpoint := os.Getenv("DB_ENDPOINT")
	dbName := os.Getenv("DB_NAME")

	// 데이터베이스 연결
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", dbUser, dbPassword, dbEndpoint, dbName))
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	defer db.Close()

	// 데이터베이스 연결 테스트
	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to database: ", err)
	}

	// 포스트 추출
	var posts []Post
	userId := c.Param("userId")
	rows, err := db.Query("SELECT post_id, user_id, image_url, content, create_at, update_at, delete_at FROM posts WHERE user_id = ?", userId)
	if err != nil {
        log.Printf("Error querying posts for user %v: %v", userId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying posts"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.PostID, &post.UserID, &post.ImageURL, &post.Content, &post.CreatedAt, &post.UpdatedAt, &post.DeletedAt); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}
		posts = append(posts, post)
	}

	// fmt.Println(posts)
	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

func GetPostByPostId(c *gin.Context){
	err := godotenv.Load("./env/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbEndpoint := os.Getenv("DB_ENDPOINT")
	dbName := os.Getenv("DB_NAME")

	// 데이터베이스 연결
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", dbUser, dbPassword, dbEndpoint, dbName))
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	defer db.Close()

	// 데이터베이스 연결 테스트
	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to database: ", err)
	}

	// 포스트 추출
	var posts []Post
	postId := c.Param("postId")
	rows, err := db.Query("SELECT post_id, user_id, image_url, content, create_at, update_at, delete_at FROM posts WHERE user_id = ?", postId)
	if err != nil {
        log.Printf("Error querying posts for user %v: %v", postId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying posts"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.PostID, &post.UserID, &post.ImageURL, &post.Content, &post.CreatedAt, &post.UpdatedAt, &post.DeletedAt); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}
		posts = append(posts, post)
	}

	// fmt.Println(posts)
	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}