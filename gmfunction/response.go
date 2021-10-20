package gmfunction

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	_ "github.com/go-sql-driver/mysql"
)

// レスポンスで返される
type Response struct {
	Payload interface{}
	Error   *ErrorResponse
}

// エラー時にステータスコードとメッセージが入る
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// レスポンスを返す
func ReplyResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	codeStr := strconv.Itoa(code)
	codeHead := codeStr[:1]
	var res Response
	if codeHead == "2" {
		res = Response{
			Payload: data,
			Error:   nil,
		}
	} else if codeHead == "4" || codeHead == "5" {
		errorRes := &ErrorResponse{
			Code:    code,
			Message: message,
		}
		res = Response{
			Payload: nil,
			Error:   errorRes,
		}
	}
	jsonData, err := json.Marshal(res)
	if err != nil {
		log.Println("json marshal error")
	}
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonData)
}