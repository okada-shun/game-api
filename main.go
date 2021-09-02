package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
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

// DataBase(game_user)からコネクション取得
func GetConnection() *gorm.DB {
	passwordBytes, err := ioutil.ReadFile("../.ssh/mysql_password")
	if err != nil {
		fmt.Println(err)
	}
	userBytes, err := ioutil.ReadFile("../.ssh/mysql_user")
	if err != nil {
		fmt.Println(err)
	}
	db, err := gorm.Open("mysql", string(userBytes)+":"+string(passwordBytes)+"@/game_user?charset=utf8&parseTime=True&loc=Local")
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
	// ID     int    `json:"id"`
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
	userId := createUUId()
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
func createUUId() string {
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
	userId := getUserId(w, r)
	if userId == "" {
		return
	}
	db := GetConnection()
	defer db.Close()
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
// トークンから名前情報を取り出し、返す
func getUserId(w http.ResponseWriter, r *http.Request) string {
	tokenString := r.Header.Get("x-token")
	token, err := VerifyToken(tokenString)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return ""
	}
	claims := token.Claims.(jwt.MapClaims)
	// claims = map[exp:1.629639808e+09 userId:bdd4056a-f113-424c-9951-1eaaaf853e5c]
	userId := claims["userId"].(string)
	return userId
}

// -H "x-token:yyy"でトークン情報を受け取り、ユーザ認証
// -d {"name":"z"}で更新する名前情報を受け取る
// トークンからユーザIDを取り出し、dbからそのユーザIDのユーザの名前情報を更新
func updateUser(w http.ResponseWriter, r *http.Request) {
	userId := getUserId(w, r)
	if userId == "" {
		return
	}
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
	// dbでnameを更新
	// UPDATE `users` SET `name` = 'Hamachan'  WHERE (user_id = 'bdd4056a-f113-424c-9951-1eaaaf853e5c')
	db.Model(&user).Where("user_id = ?", userId).Update("name", userName.Name)
	Success(w, http.StatusOK, nil)
}

type Times struct {
	Times int `json:"times"`
}

type Character struct {
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	Weight      int    `json:"weight"`
}

type UserCharacter struct {
	UserCharacterID string `json:"user_character_id"`
	UserID          string `json:"user_id"`
	CharacterID     string `json:"character_id"`
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
	userId := getUserId(w, r)
	if userId == "" {
		return
	}
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
	characters := getCharacter()
	type IDWeightSum struct {
		CharacterID string
		WeightSum   int
	}
	var idWeightSums []IDWeightSum
	arrayZero := IDWeightSum{CharacterID: "", WeightSum: 0}
	idWeightSums = append(idWeightSums, arrayZero)
	charactersCount := len(characters)
	var weightCount int
	for i := 0; i < charactersCount; i++ {
		characterId := characters[i].CharacterID
		w := characters[i].Weight
		weightCount += w
		arrayI := IDWeightSum{CharacterID: characterId, WeightSum: weightCount}
		idWeightSums = append(idWeightSums, arrayI)
	}
	var randNumbers []int
	for i := 0; i < times.Times; i++ {
		rand.Seed(time.Now().UnixNano())
		randNumber := rand.Intn(weightCount)
		randNumbers = append(randNumbers, randNumber)
	}
	type CharacterCount struct {
		CharacterID string
		Count       int
	}
	var characterCount CharacterCount
	var characterCounts []CharacterCount
	for i := 0; i < len(idWeightSums)-1; i++ {
		min := idWeightSums[i].WeightSum
		max := idWeightSums[i+1].WeightSum
		var count int
		for _, v := range randNumbers {
			if min <= v && v < max {
				count += 1
			}
		}
		characterCount = CharacterCount{CharacterID: idWeightSums[i+1].CharacterID, Count: count}
		characterCounts = append(characterCounts, characterCount)
	}
	var characterIdsDrawed []string
	for _, v := range characterCounts {
		id := v.CharacterID
		n := v.Count
		for i := 0; i < n; i++ {
			characterIdsDrawed = append(characterIdsDrawed, id)
		}
	}
	shuffle(characterIdsDrawed)
	var characterInfo CharacterResponse
	var results []CharacterResponse
	for _, character_id := range characterIdsDrawed {
		var character Character
		// SELECT * FROM `characters`  WHERE (character_id = 'c115174c-05ad-11ec-8679-a0c58933fdce')
		db.Where("character_id = ?", character_id).Find(&character)
		characterInfo = CharacterResponse{CharacterID: character_id, Name: character.Name}
		results = append(results, characterInfo)
		userCharacterId := createUUId()
		userCharacter := UserCharacter{UserCharacterID: userCharacterId, UserID: userId, CharacterID: character_id}
		/*
			INSERT INTO `user_characters` (`user_character_id`,`user_id`,`character_id`)
			VALUES ('02091c4d-1011-4615-8fbb-fd9e681153d4','703a0b0a-1541-487e-be5b-906e9541b021','c115174c-05ad-11ec-8679-a0c58933fdce')
		*/
		db.Create(&userCharacter)
	}
	Success(w, http.StatusOK, &ResultResponse{
		Results: results,
	})
	/*
		{"results":[
			{"characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Carol_N"},
			{"characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Carol_N"},
			...
			{"characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Carol_N"}
		]}
		が返る
	*/
}

// dbからcharactersテーブルの情報を取得
func getCharacter() []Character {
	db := GetConnection()
	defer db.Close()
	var character []Character
	// SELECT * FROM `characters`
	db.Find(&character)
	return character
}

// 引数の配列をシャッフルする
func shuffle(data []string) {
	n := len(data)
	for i := n - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		data[i], data[j] = data[j], data[i]
	}
}

// dbのcharactersテーブルに接続、character_idが引数character_idのデータを取得
// 取得したデータのうち、名前データを返す
func getCharacterName(character_id string) string {
	db := GetConnection()
	defer db.Close()
	var character Character
	// SELECT * FROM `characters`  WHERE (character_id = 'c115174c-05ad-11ec-8679-a0c58933fdce')
	db.Where("character_id = ?", character_id).Find(&character)
	return character.Name
}

// dbのuser_charactersテーブルに接続、user_idが引数user_idのデータを取得
func getUserCharacterList(user_id string) []UserCharacter {
	db := GetConnection()
	defer db.Close()
	var userCharacterList []UserCharacter
	// SELECT * FROM `user_characters`  WHERE (user_id = '703a0b0a-1541-487e-be5b-906e9541b021')
	db.Where("user_id = ?", user_id).Find(&userCharacterList)
	return userCharacterList
}

// localhost:8080/character/listでユーザが所持しているキャラクター一覧情報を取得
// -H "x-token:yyy"でトークン情報を受け取り、認証
func getCharacterList(w http.ResponseWriter, r *http.Request) {
	userId := getUserId(w, r)
	if userId == "" {
		return
	}
	userCharacterList := getUserCharacterList(userId)
	var characters []UserCharacterResponse
	var userCharacterInfo UserCharacterResponse
	for _, v := range userCharacterList {
		character_id := v.CharacterID
		characterName := getCharacterName(character_id)
		userCharacterInfo = UserCharacterResponse{UserCharacterID: v.UserCharacterID, CharacterID: character_id, Name: characterName}
		characters = append(characters, userCharacterInfo)
	}
	Success(w, http.StatusOK, &CharactersResponse{
		Characters: characters,
	})
	/*
		{"characters":[
			{"userCharacterID":"02091c4d-1011-4615-8fbb-fd9e681153d4","characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Carol_N"},
			{"userCharacterID":"0fed4c04-153c-4980-9a66-1424f1f7a445","characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Carol_N"},
			...
			{"userCharacterID":"95a281d5-86f0-4251-a4cb-5873231f4a96","characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Carol_N"}
		]}
		が返る
	*/
}
