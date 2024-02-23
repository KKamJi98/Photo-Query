package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"log"
	"os"
)

func ConnectDB() *sql.DB {
	// get databases environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbEndpoint := os.Getenv("DB_ENDPOINT")
	dbName := os.Getenv("DB_NAME")

	// attempt database connect
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", dbUser, dbPassword, dbEndpoint, dbName))
	if err != nil {
		log.Fatalf("Error opening database: %v", err) // error handling
	}

	// db connection test
	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	// db 객체 반환
	return db
}
