package picture

import (
	"ace-app/databases"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	// "sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// GetPictures 함수는 데이터베이스에서 모든 사진을 가져옵니다.
func GetPictures(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	var pictures []Picture

	// 데이터베이스에서 모든 사진을 조회합니다.
	rows, err := db.Query("SELECT picture_id, user_id, image_url, created_at, bookmarked FROM Pictures")
	if err != nil {
		log.Printf("사진 조회 오류: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
		return
	}
	defer rows.Close()

	// 각 행을 Picture 구조체로 스캔합니다.
	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.Bookmarked); err != nil {
			log.Printf("사진 스캔 오류: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	c.JSON(http.StatusOK, gin.H{"pictures": pictures})
}

// GetPicturesByUserId 함수는 특정 사용자 ID에 대한 사진을 가져옵니다.
func GetPicturesByUserId(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	var pictures []Picture
	userId := c.Param("user_id")

	limit := c.Query("limit")
	if limit == "" {
		limit = "20"
	}
	log.Printf("%s", limit)

	last := c.Query("last")
	if last == "" {
		last = strconv.Itoa(math.MaxInt64)
		log.Printf("%s", "last index not found")
	}
	log.Printf("last: %s", last)

	bookmark := c.Query("bookmark")
	if bookmark == "" || bookmark == "0" {
		bookmark = "0"
	} else {
		bookmark = "1"
	}

	var rows *sql.Rows
	var err error
	// 사용자 ID별로 사진을 조회합니다.
	if bookmark == "0" {
		rows, err = db.Query("SELECT picture_id, user_id, image_url, created_at, bookmarked FROM Pictures WHERE (user_id = ? AND picture_id < ?) ORDER BY picture_id DESC LIMIT ?", userId, last, limit)
		if err != nil {
			log.Printf("사용자 %v에 대한 사진 조회 오류: %v", userId, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
			return
		}
	} else {
		rows, err = db.Query("SELECT picture_id, user_id, image_url, created_at, bookmarked FROM Pictures WHERE (user_id = ? AND picture_id < ? AND bookmarked = 1) ORDER BY picture_id DESC LIMIT ?", userId, last, limit)
		if err != nil {
			log.Printf("사용자 %v에 대한 사진 조회 오류: %v", userId, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
			return
		}
	}
	defer rows.Close()

	// 각 행을 Picture 구조체로 스캔합니다.
	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.Bookmarked); err != nil {
			log.Printf("사진 스캔 오류: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	c.JSON(http.StatusOK, gin.H{"pictures": pictures})
}

// GetPictureByPictureId 함수는 ID로 단일 사진을 가져옵니다.
func GetPictureByPictureId(c *gin.Context) {
	db := database.ConnectDB()
	defer db.Close()

	var pictures []Picture
	pictureId := c.Param("picture_id")

	// 사진 ID로 사진을 조회합니다.
	rows, err := db.Query("SELECT picture_id, user_id, image_url, created_at, bookmarked FROM Pictures WHERE picture_id = ?", pictureId)
	if err != nil {
		log.Printf("사진 조회 오류: %v, %v", pictureId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
		return
	}
	defer rows.Close()

	// 각 행을 Picture 구조체로 스캔합니다.
	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.Bookmarked); err != nil {
			log.Printf("사진 스캔 오류: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}

	c.JSON(http.StatusOK, gin.H{"pictures": pictures})
}

func GetPicturesByTags(c *gin.Context) {
	/**
	1. Query Paramerter와 Path Parameter로 tag와 user_id 받기 [x]
	2. DynamoDB에서 user_id로 조회 [v]
	3. DynamoDB에서 쿼리파라미터로 들어온 태그가 존재하는 것만 선택후 배열에 저장
	3. uploadtime으로 내림차순으로 정렬
	4. limit, lastindex로 무한스크롤 구현 (슬라이스 사용)
	4. 반환
	*/

	tagFilter := c.Query("tag")
	if tagFilter == "" {
		log.Printf("%v", "no tag specified")
	}
	log.Printf("tag => %v", tagFilter)

	userId := c.Param("user_id")
	if userId == "" {
		log.Printf("user_id => %v", "no user_id specified")
	}
	log.Printf("%v", userId)

	limit := c.Query("limit")
	if limit == "" {
		limit = "20"
	}
	log.Printf("limit => %s", limit)
	// last = strconv.Itoa(0)

	last := c.Query("last")
	if last == "" {
		last = strconv.Itoa(0)
		// last = strconv.Itoa(math.MaxInt64)
		log.Printf("%s", "last index not found")
	}
	log.Printf("last=> %s", last)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}
	dynamoDbClient := dynamodb.NewFromConfig(cfg)

	tableBasics := TableBasics{
		DynamoDbClient: dynamoDbClient,
		TableName:      os.Getenv("DYNAMODB_TABLE"),
	}
	log.Println(tableBasics.TableName, userId)
	// Query API 호출
	response, err := tableBasics.DynamoDbClient.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(tableBasics.TableName),
		KeyConditionExpression: aws.String("user_id = :userId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: userId},
		},
	})
	if err != nil {
		log.Printf("Couldn't query the table. Here's why: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DynamoDB 테이블 조회 오류"})
	}

	var tags []Tag
	// 응답에서 태그 목록 언마샬링
	err = attributevalue.UnmarshalListOfMaps(response.Items, &tags)
	if err != nil {
		log.Printf("Couldn't unmarshal response items. Here's why: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DynamoDB item un-marshaling 오류"})
		return
	}

	var selectedTags []Tag
	for _, tag := range tags {
		for _, Value := range tag.Tags {
			if Value == tagFilter {
				selectedTags = append(selectedTags, tag)
			}
		}
	}

	placeholders := make([]string, len(selectedTags))
	params := make([]interface{}, len(selectedTags))
	for i, id := range selectedTags {
		placeholders[i] = "?"
		params[i] = id.ImageId
	}
	inClause := strings.Join(placeholders, ",")

	db := database.ConnectDB()
	defer db.Close()

	query := fmt.Sprintf("SELECT picture_id, user_id, image_url, created_at, bookmarked FROM Pictures WHERE image_url IN (%s)", inClause)

	rows, err := db.Query(query, params...)
	if err != nil {
		log.Printf("사진 조회 오류: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사진 조회 오류"})
		return
	}
	defer rows.Close()

	var pictures []Picture
	for rows.Next() {
		var picture Picture
		if err := rows.Scan(&picture.PictureID, &picture.UserID, &picture.ImageURL, &picture.CreatedAt, &picture.Bookmarked); err != nil {
			log.Printf("사진 스캔 오류: %v", err)
			continue
		}
		pictures = append(pictures, picture)
	}
	// sort.Slice(selectedTags, func(i, j int) bool {
	// 	return selectedTags[i].UploadTime > selectedTags[j].UploadTime
	// })

	lastIndex, err := strconv.ParseInt(last, 10, 64)
	if err != nil {
		// 에러 처리: last 값을 파싱할 수 없음
		log.Printf("Error parsing last index: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid last index"})
		return
	}

	limitInt, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		// 에러 처리: limit 값을 파싱할 수 없음
		log.Printf("Error parsing limit: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}
	startIndex := int(lastIndex)
	endIndex := startIndex + int(limitInt)

	// endIndex가 슬라이스 길이를 초과하지 않도록 조정
	if endIndex > len(pictures) {
		endIndex = len(pictures)
	}

	log.Printf("last => %v, end => %v len => %v", startIndex, endIndex, len(pictures))

	// 슬라이스 범위 확인
	if startIndex < 0 || endIndex > len(selectedTags) || startIndex >= len(pictures) {
		// 범위가 유효하지 않은 경우의 처리
		c.JSON(http.StatusBadRequest, gin.H{"error": "Index out of range"})
		return
	}

	// 선택된 태그의 부분 슬라이스 반환
	c.JSON(http.StatusOK, gin.H{"pictures": pictures[startIndex:endIndex]})
}
