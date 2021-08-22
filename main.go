package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Router作成
	router := mux.NewRouter()
	// URLと処理
	router.HandleFunc("/", home)
	// ユーザ関連API
	router.HandleFunc("/user/create", createUser).Methods("POST")
	router.HandleFunc("/user/get", getUser).Methods("GET")
	router.HandleFunc("/user/update", updateUser).Methods("PUT")
	// ガチャ関連API
	router.HandleFunc("/gacha/draw", drawGacha).Methods("POST")
	// キャラクター関連API
	router.HandleFunc("/character/list", getCharacterList).Methods("GET")
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
		log.Printf("DB connection failed %v", err)
	}
	db.LogMode(true)
	return db
}

type UserName struct {
	Name string `json:"name"`
}

type User struct {
	ID     int    `json:"id"`
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

// localhost:8080/user/createでユーザ情報を作成
// -d {"name":"x"}で名前情報を受け取る
// UUIDでユーザIDを生成する
// ユーザIDからjwtでトークンを作成し、トークンを返す
func createUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	// string(body) = {"name": "x"}
	var userName UserName
	if err := json.Unmarshal(body, &userName); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		return
	}
	userId := createId()
	user := User{UserID: userId, Name: userName.Name}
	db := GetConnection()
	defer db.Close()
	// INSERT INTO `users` (`user_id`,`name`) VALUES ('bdd4056a-f113-424c-9951-1eaaaf853e5c','Tamachan')
	db.Create(&user)
	// ユーザIDの文字列からjwtでトークン作成
	token, err := CreateToken(userId)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	// token = "生成されたトークンの文字列"
	Success(w, http.StatusOK, &TokenResponse{
		Token: token,
	})
	// {"token":"生成されたトークンの文字列"}が返る
}

// ユーザIDからjwtでトークンを作成
// 有効期限は24時間に設定
// jwtのペイロードにはユーザIDと有効期限の時刻を設定
func CreateToken(userID string) (string, error) {
	// HS256は256ビットのハッシュ値を生成するアルゴリズム
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	// ペイロードにユーザIDと有効期限の時刻(24時間後)を設定
	token.Claims = jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	}
	// 秘密鍵を取得
	signBytes, err := ioutil.ReadFile("../.ssh/id_rsa")
	if err != nil {
		fmt.Println(err)
	}
	// 秘密鍵で署名
	tokenString, err := token.SignedString(signBytes)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// UUIDでユーザIDを生成
func createId() string {
	u, err := uuid.NewRandom()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	uu := u.String()
	return uu
}

func Success(w http.ResponseWriter, code int, data interface{}) {
	jsonData, _ := json.Marshal(data)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
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
		log.Println("json marshal error")
	}
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonData)
}

// getUser関数で返されるユーザの名前情報が入る
type UserResponse struct {
	Name string `json:"name"`
}

// -H "x-token:yyy"でトークン情報を受け取り、ユーザ認証
// トークンからユーザIDを取り出し、dbからそのユーザIDのユーザの名前情報を取り出し、返す
func getUser(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("x-token")
	// tokenString = yyy
	// VerifyToken関数で認証
	token, err := VerifyToken(tokenString)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	/*例
	token = &{eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.
		eyJleHAiOjE2Mjk2Mzk4MDgsInVzZXJJZCI6ImJkZDQwNTZhLWYxMTMtNDI0Yy05OTUxLTFlYWFhZjg1M2U1YyJ9.
		xlSV0fVPtzWfjGT7GQc5sACnrS2R4T4B-ivxq15eagc 0xc00000e0c0
		map[alg:HS256 typ:JWT] map[exp:1.629639808e+09 userId:bdd4056a-f113-424c-9951-1eaaaf853e5c]
		xlSV0fVPtzWfjGT7GQc5sACnrS2R4T4B-ivxq15eagc true}
	*/
	db := GetConnection()
	defer db.Close()
	claims := token.Claims.(jwt.MapClaims)
	// claims = map[exp:1.629639808e+09 userId:bdd4056a-f113-424c-9951-1eaaaf853e5c]
	userId := claims["userId"]
	// userId = bdd4056a-f113-424c-9951-1eaaaf853e5c
	var user User
	// SELECT * FROM `users`  WHERE (user_id = 'bdd4056a-f113-424c-9951-1eaaaf853e5c')
	db.Where("user_id = ?", userId).Find(&user)
	Success(w, http.StatusOK, &UserResponse{
		Name: user.Name,
	})
	// {"name":"x"}が返る
	// 有効期限が切れると{"code":400,"message":"Token is expired"}が返る
}

// jwtトークンを認証する
func VerifyToken(tokenString string) (*jwt.Token, error) {
	// 秘密鍵を取得
	signBytes, err := ioutil.ReadFile("../.ssh/id_rsa")
	if err != nil {
		fmt.Println(err)
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signBytes, nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// -H "x-token:yyy"でトークン情報を受け取り、ユーザ認証
// -d {"name":"z"}で更新する名前情報を受け取る
// トークンからユーザIDを取り出し、dbからそのユーザIDのユーザの名前情報を更新
func updateUser(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("x-token")
	// tokenString = yyy
	// VerifyToken関数で認証
	token, err := VerifyToken(tokenString)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	/*例
	token = &{eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.
		eyJleHAiOjE2Mjk2Mzk4MDgsInVzZXJJZCI6ImJkZDQwNTZhLWYxMTMtNDI0Yy05OTUxLTFlYWFhZjg1M2U1YyJ9.
		xlSV0fVPtzWfjGT7GQc5sACnrS2R4T4B-ivxq15eagc 0xc00000e0c0
		map[alg:HS256 typ:JWT] map[exp:1.629639808e+09 userId:bdd4056a-f113-424c-9951-1eaaaf853e5c]
		xlSV0fVPtzWfjGT7GQc5sACnrS2R4T4B-ivxq15eagc true}
	*/
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	// string(body) = {"name": "z"}
	var userName UserName
	if err := json.Unmarshal(body, &userName); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		return
	}
	var user User
	db := GetConnection()
	defer db.Close()
	claims := token.Claims.(jwt.MapClaims)
	// claims = map[exp:1.629639808e+09 userId:bdd4056a-f113-424c-9951-1eaaaf853e5c]
	userId := claims["userId"]
	// userId = bdd4056a-f113-424c-9951-1eaaaf853e5c
	// dbでnameを更新
	// UPDATE `users` SET `name` = 'Hamachan'  WHERE (user_id = 'bdd4056a-f113-424c-9951-1eaaaf853e5c')
	db.Model(&user).Where("user_id = ?", userId).Update("name", userName.Name)
	Success(w, http.StatusOK, "")
}

type Times struct {
	Times int `json:"times"`
}

type Rarity struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}

type Character struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Rarity string `json:"rarity"`
}

type UserCharacter struct {
	ID          int    `json:"id"`
	UserID      string `json:"user_id"`
	CharacterID string `json:"character_id"`
}

type CharacterResponse struct {
	CharacterID string `json:"characterID"`
	Name        string `json:"name"`
}

// drawGacha関数で返される
type ResultResponse struct {
	Results []CharacterResponse `json:"results"`
}

type UserCharacterResponse struct {
	UserCharacterID string `json:"userCharacterID"`
	CharacterID     string `json:"characterID"`
	Name            string `json:"name"`
}

// getCharacterList関数で返される
type CharactersResponse struct {
	Characters []UserCharacterResponse `json:"characters"`
}

// localhost:8080/gacha/drawでガチャを引いて、キャラクターを取得
// -H "x-token:yyy"でトークン情報を受け取り、認証
// -d {"times":x}でガチャを何回引くかの情報を受け取る
func drawGacha(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("x-token")
	token, err := VerifyToken(tokenString)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	id := claims["userId"].(string)
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	var times Times
	if err := json.Unmarshal(body, &times); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		return
	}
	db := GetConnection()
	defer db.Close()
	rarityValue := getRarity()
	maxA := rarityValue[0].Weight
	maxB := rarityValue[1].Weight
	maxOne := maxA - 1
	maxTwo := maxA + maxB - 1
	// 引いたキャラのキャラクターIDが入る
	var characterIdList []int
	// 引いたキャラのキャラクターIDとキャラ名が入る
	var characterInfo CharacterResponse
	// 引いたキャラのキャラクターIDとキャラ名が、引いたキャラの数だけ入る
	var results []CharacterResponse
	// 0から99までの整数をランダムにtimes.Times回数分だけ生成
	// 0以上maxOne未満でレア度が1のキャラを引き、maxOne以上maxTwo未満でレア度が2のキャラを引き、
	// maxTwo以上99以下でレア度が3のキャラを引く
	for i := 0; i < times.Times; i++ {
		rand.Seed(time.Now().UnixNano())
		randNumber := rand.Intn(100)
		var character Character
		if randNumber < maxOne {
			// SELECT * FROM `characters`  WHERE (rarity = 1)
			db.Where("rarity = 1").Find(&character)
			characterIdList = append(characterIdList, character.ID)
			characterInfo = CharacterResponse{CharacterID: strconv.Itoa(character.ID), Name: character.Name}
			results = append(results, characterInfo)
		} else if randNumber < maxTwo {
			// SELECT * FROM `characters`  WHERE (rarity = 2)
			db.Where("rarity = 2").Find(&character)
			characterIdList = append(characterIdList, character.ID)
			characterInfo = CharacterResponse{CharacterID: strconv.Itoa(character.ID), Name: character.Name}
			results = append(results, characterInfo)
		} else {
			// SELECT * FROM `characters`  WHERE (rarity = 3)
			db.Where("rarity = 3").Find(&character)
			characterIdList = append(characterIdList, character.ID)
			characterInfo = CharacterResponse{CharacterID: strconv.Itoa(character.ID), Name: character.Name}
			results = append(results, characterInfo)
		}
	}
	// 引いたキャラのキャラクターIDと、そのキャラの名前をUserCharacter構造体に入れ、
	// dbのuser_charactersテーブルに格納していく
	for _, v := range characterIdList {
		userCharacter := UserCharacter{UserID: id, CharacterID: strconv.Itoa(v)}
		db.Create(&userCharacter)
	}
	Success(w, http.StatusOK, &ResultResponse{
		Results: results,
	})
	/*
		{"results":[
			{"characterID":"3","name":"Carol"},
			{"characterID":"3","name":"Carol"},
			...
			{"characterID":"3","name":"Carol"}
		]}
		が返る
	*/
}

// dbのraritiesテーブルに接続、全データを取得
// SELECT * FROM `rarities`
func getRarity() []Rarity {
	db := GetConnection()
	defer db.Close()
	var rarity []Rarity
	db.Find(&rarity)
	return rarity
}

// dbのcharactersテーブルに接続、idが引数idのデータを取得
// SELECT * FROM `characters`  WHERE (id = 1)
// 取得したデータのうち、名前データを返す
func getCharacterName(id int) string {
	db := GetConnection()
	defer db.Close()
	var character Character
	db.Where("id = ?", id).Find(&character)
	return character.Name
}

// dbのuser_charactersテーブルに接続、user_idが引数user_idのデータを取得
// SELECT * FROM `user_characters`  WHERE (user_id = '1')
func getUserCharacterList(user_id string) []UserCharacter {
	db := GetConnection()
	defer db.Close()
	var userCharacterList []UserCharacter
	db.Where("user_id = ?", user_id).Find(&userCharacterList)
	return userCharacterList
}

// localhost:8080/character/listでユーザが所持しているキャラクター一覧情報を取得
// -H "x-token:yyy"でトークン情報を受け取り、認証
func getCharacterList(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("x-token")
	token, err := VerifyToken(tokenString)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	id := claims["userId"].(string)
	userCharacterList := getUserCharacterList(id)
	var characters []UserCharacterResponse
	var userCharacterInfo UserCharacterResponse
	for _, v := range userCharacterList {
		id, _ := strconv.Atoi(v.CharacterID)
		characterName := getCharacterName(id)
		userCharacterInfo = UserCharacterResponse{UserCharacterID: strconv.Itoa(v.ID), CharacterID: v.CharacterID, Name: characterName}
		characters = append(characters, userCharacterInfo)
	}
	Success(w, http.StatusOK, &CharactersResponse{
		Characters: characters,
	})
	/*
		{"characters":[
			{"userCharacterID":"1","characterID":"3","name":"Carol"},
			{"userCharacterID":"2","characterID":"3","name":"Carol"},
			...
			{"userCharacterID":"100","characterID":"3","name":"Carol"}
		]}
		が返る
	*/
}
