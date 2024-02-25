package picture

import (
	"context"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
)

// Tag 구조체 정의
type Tag struct {
	UserId  string   `dynamodbav:"user_id"`
	ImageId string   `dynamodbav:"images_id"`
	Tags    []string `dynamodbav:"tags"`
}

type TableBasics struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}

func GetAllTags(c *gin.Context) {
	// 프론트엔드에서 넘어온 userId 값 받기
	userId := c.Query("user_id") // URL 파라미터에서 user_id 값을 읽음

	// AWS 설정 로드 및 DynamoDB 클라이언트 초기화
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}
	dynamoDbClient := dynamodb.NewFromConfig(cfg)

	tableBasics := TableBasics{
		DynamoDbClient: dynamoDbClient,
		TableName:      os.Getenv("DYNAMODB_TABLE"),
	}

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

	GetRank(c, response)
}

func GetRank(c *gin.Context, data *dynamodb.QueryOutput) {
	var tags []Tag

	// 응답에서 태그 목록 언마샬링
	err := attributevalue.UnmarshalListOfMaps(data.Items, &tags)
	if err != nil {
		log.Printf("Couldn't unmarshal response items. Here's why: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DynamoDB item un-marshaling 오류"})
		return
	}

	// 태그의 등장 횟수를 계산
	tagCounts := make(map[string]int)
	for _, tag := range tags {
		for _, t := range tag.Tags {
			tagCounts[t]++
		}
	}

	// 태그와 등장 횟수를 저장할 슬라이스 생성
	type tagCount struct {
		Tag   string
		Count int
	}
	var sortedTags []tagCount
	for tag, count := range tagCounts {
		sortedTags = append(sortedTags, tagCount{Tag: tag, Count: count})
	}

	// 등장 횟수에 따라 슬라이스 정렬 (내림차순)
	sort.Slice(sortedTags, func(i, j int) bool {
		return sortedTags[i].Count > sortedTags[j].Count
	})

	// 상위 5개 태그 선택
	var topTags []tagCount
	if len(sortedTags) > 5 {
		topTags = sortedTags[:5]
	} else {
		topTags = sortedTags
	}

	// 등장 횟수에 따른 상위 5개 태그 반환
	c.JSON(http.StatusOK, gin.H{"topTags": topTags})
}
