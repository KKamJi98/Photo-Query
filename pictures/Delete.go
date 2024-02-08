package picture

import (
	"ace-app/databases"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// DeletePicturesByPostId는 그림들을 ID로 삭제하는 작업을 처리합니다.
func DeletePicturesByPostId(c *gin.Context) {

	// 예상되는 JSON 데이터 형식과 일치하는 구조체를 정의합니다.
	// type picture struct {
	// 	PictureID int64 `json:"picture_id"`
	// }

	var pictures []Picture

	// 요청에서 JSON 데이터를 추출하고 언마샬합니다.
	jsonData := c.PostForm("json_data")
	log.Printf("%v \t %T", jsonData, jsonData)
	if err := json.Unmarshal([]byte(jsonData), &pictures); err != nil {
		log.Printf("잘못된 JSON 데이터: %v", err)
		c.JSON(400, gin.H{"message": "잘못된 JSON 데이터", "error": err.Error()})
		return
	}

	db := database.ConnectDB()
	defer db.Close()

	// 삭제 작업 결과를 추적하기 위한 카운터를 초기화합니다.
	errorCount := 0
	successCount := 0
	nofoundCount := 0

	// 각 그림 ID를 반복하고 삭제를 시도합니다.
	for _, pic := range pictures {
		result, err := db.Exec("DELETE FROM Pictures WHERE picture_id = ?", pic.PictureID)
		if err != nil {
			log.Printf("%d번 ID를 가진 그림 삭제 중 오류 발생: %v", pic.PictureID, err)
			errorCount++
			continue
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("%d번 ID에 대한 영향 받는 행 수를 가져오는 중 오류 발생: %v", pic.PictureID, err)
			errorCount++
			continue
		}

		if rowsAffected == 0 {
			log.Printf("%d번 ID를 가진 그림을 찾을 수 없음", pic.PictureID)
			nofoundCount++
		} else {
			successCount++
		}
	}

	// 작업 결과를 로그에 기록합니다.
	log.Printf("%d개의 그림 삭제 || %d개의 오류 발생 || %d개의 그림 ID를 찾을 수 없음", successCount, errorCount, nofoundCount)

	// 삭제 작업 결과를 응답합니다.
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d개의 그림 삭제, %d개의 오류 발생, %d개의 그림 ID를 찾을 수 없음", successCount, errorCount, nofoundCount)})
}
