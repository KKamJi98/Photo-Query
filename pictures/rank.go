package picture

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

// Tag 구조체 정의
type Tag struct {
	UserId  string `dynamodbav:"user_id"`
	ImageId string `dynamodbav:"images_id"`
	Tags    []string `dynamodbav:"tags"`
}

type TableBasics struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}

func GetAllTags(c *gin.Context) {
	//aws 설정 로드 및 DynamoDB 클라이언트 초기화
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1")) // 필요에 따라 리전 변경
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}
	dynamoDbClient := dynamodb.NewFromConfig(cfg)

	tableBasics := TableBasics{
		DynamoDbClient: dynamoDbClient,
		TableName:      os.Getenv("DYNAMODB_TABLE"), // 환경 변수에서 DynamoDB 테이블 이름 로드
	}
	var tags []Tag

	// Scan API 호출
	response, err := tableBasics.DynamoDbClient.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(tableBasics.TableName),
	})

	if err != nil {
		log.Printf("Couldn't scan the table. Here's why: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DynamoDB 테이블 조회 오류"})
		return
	}

	// 응답에서 태그 목록 언마샬링
	err = attributevalue.UnmarshalListOfMaps(response.Items, &tags)
	if err != nil {
		log.Printf("Couldn't unmarshal response items. Here's why: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DynamoDB item un-marshaling 오류"})
		return
	}

	for _, tag := range tags {
		log.Println(tag)
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}
