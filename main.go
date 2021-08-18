package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Router作成
	router := mux.NewRouter().StrictSlash(true)
	// URLと処理
	router.HandleFunc("/", home)
	router.HandleFunc("/user/create", createUser).Methods("POST")
	router.HandleFunc("/user/get", getUser).Methods("GET")
	router.HandleFunc("/user/update", updateUser).Methods("PUT")
	// ポートを8080で指定してRouter起動
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Hello Worldをlocalhost:8080画面に表示（後に削除予定）
func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

// Error情報をJsonで返す
func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, map[string]string{"error": msg})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// DataBase(game_user)と接続
func GetConnection() *gorm.DB {
	db, err := gorm.Open("mysql", "okada:password@/game_user?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatalf("DB connection failed %v", err)
	}
	db.LogMode(true)
	return db
}

// User構造体
type User struct {
	ID   int
	Name string
}

// createUser関数で作成されたjwtトークンが入る
type LoginResponse struct {
	Token string `json:"token"`
}

// localhost:8080/user/createでユーザ情報を作成
// -d {"name":"　"}で名前情報を受け取り、dbに挿入
// dbのIDからjwtトークンを作成し、トークンを返す
func createUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		return
	}
	db := GetConnection()
	defer db.Close()
	db.Create(&user)
	token, err := CreateToken(strconv.Itoa(user.ID))
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	Success(w, &LoginResponse{
		Token: token,
	})
}

// ユーザIDからjwtトークンを作成
// 期限は1時間
func CreateToken(userID string) (string, error) {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims = jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(time.Hour * 1).Unix(),
	}
	var secretKey = "secret"
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func Success(w http.ResponseWriter, data interface{}) {
	jsonData, _ := json.Marshal(data)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonData)
}

// Errorメッセージが入る
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Errorメッセージを返す
func ErrorResponse(w http.ResponseWriter, code int, message string) {
	jsonData, err := json.Marshal(&Response{
		Code:    code,
		Message: message,
	})
	if err != nil {
		log.Fatal("json marshal error")
	}
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonData)
}

// getUser関数で返されるユーザの名前情報が入る
type UserResponse struct {
	Name string `json:"name"`
}

// -H "x-token: "でトークン情報を受け取り、認証
// トークンからユーザIDを取り出し、dbからそのIDのユーザの名前情報を取り出し、返す
func getUser(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("x-token")
	token, err := VerifyToken(tokenString)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	db := GetConnection()
	defer db.Close()
	claims := token.Claims.(jwt.MapClaims)
	id := claims["userId"]
	var user User
	db.Where("id = ?", id).Find(&user)
	Success(w, &UserResponse{
		Name: user.Name,
	})
}

// jwtトークンを認証する
func VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// -H "x-token: "でトークン情報を受け取り、認証
// -d {"name":" "}で更新する名前情報を受け取る
// トークンからユーザIDを取り出し、dbからそのIDのユーザの名前情報を更新
func updateUser(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("x-token")
	token, err := VerifyToken(tokenString)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		return
	}
	db := GetConnection()
	defer db.Close()
	claims := token.Claims.(jwt.MapClaims)
	id := claims["userId"]
	db.Model(&user).Where("id = ?", id).Update("name", user.Name)
}
