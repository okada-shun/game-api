package gmfunction

import (
	"encoding/json"
	"log"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
)

// エラー時にステータスコードとメッセージが入る
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string   `json:"status"`
}

// エラーレスポンスを返す
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, ErrorResponse{Code: code, Message: message, Status: "Error"})
}

// レスポンスを返す
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Println("json marshal error")
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}