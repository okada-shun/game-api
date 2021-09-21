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
	wr "github.com/mroth/weightedrand"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

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
	db, err := gorm.Open(mysql.Open(string(userBytes)+":"+string(passwordBytes)+"@/game_user?charset=utf8&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		log.Printf("DB connection failed %v", err)
	}
	db.Logger = db.Logger.LogMode(logger.Info)
	return db
}

type UserName struct {
	Name string `json:"name"`
}

type User struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

// localhost:8080/user/createでユーザ情報を作成
// -d {"name":"aaa"}で名前情報を受け取る
// UUIDでユーザIDを生成する
// ユーザIDからjwtでトークンを作成し、トークンを返す
func createUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	// string(body) = {"name": "aaa"}
	var userName UserName
	if err := json.Unmarshal(body, &userName); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		return
	}
	userId := createUUId()
	user := User{UserID: userId, Name: userName.Name}
	db := GetConnection()
	db_sql, err := db.DB()
	if err != nil {
		fmt.Println(err)
	}
	defer db_sql.Close()
	// INSERT INTO `users` (`user_id`,`name`) VALUES ('bdd4056a-f113-424c-9951-1eaaaf853e5c','aaa')
	db.Create(&user)
	// ユーザIDの文字列からjwtでトークン作成
	token, err := createToken(userId)
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
func createToken(userID string) (string, error) {
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
	db_sql, err := db.DB()
	if err != nil {
		fmt.Println(err)
	}
	defer db_sql.Close()
	var user User
	// SELECT * FROM `users`  WHERE (user_id = 'bdd4056a-f113-424c-9951-1eaaaf853e5c')
	db.Where("user_id = ?", userId).Find(&user)
	Success(w, http.StatusOK, &UserResponse{
		Name: user.Name,
	})
	// {"name":"aaa"}が返る
	// 有効期限が切れると{"code":400,"message":"Token is expired"}が返る
}

// jwtトークンを認証する
func verifyToken(tokenString string) (*jwt.Token, error) {
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
	token, err := verifyToken(tokenString)
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
// -d {"name":"bbb"}で更新する名前情報を受け取る
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
	// string(body) = {"name": "bbb"}
	var userName UserName
	if err := json.Unmarshal(body, &userName); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		return
	}
	var user User
	db := GetConnection()
	db_sql, err := db.DB()
	if err != nil {
		fmt.Println(err)
	}
	defer db_sql.Close()
	// dbでnameを更新
	// UPDATE `users` SET `name` = 'bbb'  WHERE (user_id = 'bdd4056a-f113-424c-9951-1eaaaf853e5c')
	db.Model(&user).Where("user_id = ?", userId).Update("name", userName.Name)
	Success(w, http.StatusOK, nil)
}

type Times struct {
	Times int `json:"times"`
}

type GachaID struct {
	GachaID int `json:"gacha_id"`
}

type Character struct {
	CharacterID   string `json:"character_id"`
	CharacterName string `json:"character_name"`
	Weight        uint   `json:"weight"`
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
// -d {"gacha_id":n, "times":x}でどのガチャを引くか、ガチャを何回引くかの情報を受け取る
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
	var gachaId GachaID
	if err := json.Unmarshal(body, &gachaId); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		return
	}
	var times Times
	if err := json.Unmarshal(body, &times); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		return
	}
	db := GetConnection()
	db_sql, err := db.DB()
	if err != nil {
		fmt.Println(err)
	}
	defer db_sql.Close()
	charactersList := getCharacters(gachaId.GachaID)
	characterIdsDrawed := drawCharacterIds(charactersList, times.Times)
	var characterInfo CharacterResponse
	var results []CharacterResponse
	var userCharacters []UserCharacter
	count := 0
	for _, character_id := range characterIdsDrawed {
		character := getCharacterInfo(charactersList, character_id)
		characterInfo = CharacterResponse{CharacterID: character_id, Name: character.CharacterName}
		results = append(results, characterInfo)
		userCharacterId := createUUId()
		userCharacter := UserCharacter{UserCharacterID: userCharacterId, UserID: userId, CharacterID: character_id}
		userCharacters = append(userCharacters, userCharacter)
		count += 1
		if count == 10000 {
			/*
				INSERT INTO `user_characters` (`user_character_id`,`user_id`,`character_id`)
				VALUES ('eaaada0c-3815-4da2-b791-3447a816a3e0','c2f0d74b-0321-4f87-930f-8d85350ee6d4','7b6a8a4e-0ed8-11ec-93f3-a0c58933fdce')
				, ... ,
				('ff1583af-3f60-43de-839c-68094286e11a','c2f0d74b-0321-4f87-930f-8d85350ee6d4','7b6d0b6d-0ed8-11ec-93f3-a0c58933fdce')
			*/
			db.Create(&userCharacters)
			userCharacters = userCharacters[:0]
			count = 0
		}
	}
	if len(userCharacters) != 0 {
		/*
			INSERT INTO `user_characters` (`user_character_id`,`user_id`,`character_id`)
			VALUES ('98b27372-8806-4d33-950a-68625ed6d687','c2f0d74b-0321-4f87-930f-8d85350ee6d4','7b6c0f26-0ed8-11ec-93f3-a0c58933fdce')
		*/
		db.Create(&userCharacters)
	}
	Success(w, http.StatusOK, &ResultResponse{
		Results: results,
	})
	/*
		{"results":[
			{"characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Sun"},
			{"characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Venus"},
			...
			{"characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Pluto"}
		]}
		が返る
	*/
}

// charactersListからキャラクターのidとweightを取り出しchoicesに格納
// times回分だけchoicesからWeighted Random Selectionを実行
func drawCharacterIds(charactersList []Character, times int) []string {
	var choices []wr.Choice
	for i := 0; i < len(charactersList); i++ {
		choices = append(choices, wr.Choice{Item: charactersList[i].CharacterID, Weight: charactersList[i].Weight})
	}
	rand.Seed(time.Now().UTC().UnixNano())
	var characterIdsDrawed []string
	for i := 0; i < times; i++ {
		chooser, _ := wr.NewChooser(choices...)
		characterIdsDrawed = append(characterIdsDrawed, chooser.Pick().(string))
	}
	return characterIdsDrawed
}

// dbからキャラクターのid、名前、weightの情報を取得
// ガチャidが引数gacha_idのキャラクターに限る
func getCharacters(gacha_id int) []Character {
	db := GetConnection()
	db_sql, err := db.DB()
	if err != nil {
		fmt.Println(err)
	}
	defer db_sql.Close()
	var charactersList []Character
	/*
		SELECT gacha_characters.character_id, gacha_characters.character_name, rarities.weight
		FROM `gacha_characters`
		join rarities
		on gacha_characters.rarity_id = rarities.rarity_id
		WHERE gacha_id = 1
	*/
	db.Table("gacha_characters").Select("gacha_characters.character_id, gacha_characters.character_name, rarities.weight").Joins("join rarities on gacha_characters.rarity_id = rarities.rarity_id").Where("gacha_id = ?", gacha_id).Scan(&charactersList)
	return charactersList
}

// dbから全てのキャラクターのid、名前、weightの情報を取得
func getAllCharacters() []Character {
	db := GetConnection()
	db_sql, err := db.DB()
	if err != nil {
		fmt.Println(err)
	}
	defer db_sql.Close()
	var charactersList []Character
	/*
		SELECT gacha_characters.character_id, gacha_characters.character_name, rarities.weight
		FROM `gacha_characters`
		join rarities
		on gacha_characters.rarity_id = rarities.rarity_id
	*/
	db.Table("gacha_characters").Select("gacha_characters.character_id, gacha_characters.character_name, rarities.weight").Joins("join rarities on gacha_characters.rarity_id = rarities.rarity_id").Scan(&charactersList)
	return charactersList
}

// 引数のcharactersListからCharacterIDが引数character_idのデータを取得
func getCharacterInfo(charactersList []Character, character_id string) Character {
	for i := 0; i < len(charactersList); i++ {
		if charactersList[i].CharacterID == character_id {
			return charactersList[i]
		}
	}
	return Character{}
}

// dbのuser_charactersテーブルからuser_idが引数user_idのデータを取得
func getUserCharacters(user_id string) []UserCharacter {
	db := GetConnection()
	db_sql, err := db.DB()
	if err != nil {
		fmt.Println(err)
	}
	defer db_sql.Close()
	var userCharactersList []UserCharacter
	// SELECT * FROM `user_characters`  WHERE (user_id = '703a0b0a-1541-487e-be5b-906e9541b021')
	db.Where("user_id = ?", user_id).Find(&userCharactersList)
	return userCharactersList
}

// localhost:8080/character/listでユーザが所持しているキャラクター一覧情報を取得
// -H "x-token:yyy"でトークン情報を受け取り、認証
func getCharacterList(w http.ResponseWriter, r *http.Request) {
	userId := getUserId(w, r)
	if userId == "" {
		return
	}
	allCharactersList := getAllCharacters()
	userCharactersList := getUserCharacters(userId)
	var characters []UserCharacterResponse
	var userCharacterInfo UserCharacterResponse
	if len(userCharactersList) == 0 {
		characters = make([]UserCharacterResponse, 0)
	} else {
		for _, v := range userCharactersList {
			character_id := v.CharacterID
			character := getCharacterInfo(allCharactersList, character_id)
			characterName := character.CharacterName
			userCharacterInfo = UserCharacterResponse{UserCharacterID: v.UserCharacterID, CharacterID: character_id, Name: characterName}
			characters = append(characters, userCharacterInfo)
		}
	}
	Success(w, http.StatusOK, &CharactersResponse{
		Characters: characters,
	})
	/*
		{"characters":[
			{"userCharacterID":"02091c4d-1011-4615-8fbb-fd9e681153d4","characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Sun"},
			{"userCharacterID":"0fed4c04-153c-4980-9a66-1424f1f7a445","characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Venus"},
			...
			{"userCharacterID":"95a281d5-86f0-4251-a4cb-5873231f4a96","characterID":"c115174c-05ad-11ec-8679-a0c58933fdce","name":"Pluto"}
		]}
		が返る
	*/
}
