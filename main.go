package main

import (
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
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

// DataBase(game_user)からコネクション取得
func GetConnection() (*gorm.DB, error) {
	passwordBytes, err := ioutil.ReadFile("../.ssh/mysql_password")
	if err != nil {
		return nil, err
	}
	userBytes, err := ioutil.ReadFile("../.ssh/mysql_user")
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(mysql.Open(string(userBytes)+":"+string(passwordBytes)+"@/game_user?charset=utf8&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.Logger = db.Logger.LogMode(logger.Info)
	return db, nil
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
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	userId, err := createUUId()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	user.UserID = userId

	db, err := GetConnection()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	db_sql, err := db.DB()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer db_sql.Close()
	/*
		INSERT INTO `users` (`user_id`,`name`) VALUES ('95daec2b-287c-4358-ba6f-5c29e1c3cbdf','aaa')
	*/
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
		return "", err
	}
	// 秘密鍵で署名
	tokenString, err := token.SignedString(signBytes)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// UUIDでユーザIDを生成
func createUUId() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	uu := u.String()
	return uu, nil
}

// getUser関数で返されるユーザの名前情報が入る
type UserResponse struct {
	Name string `json:"name"`
}

// -H "x-token:yyy"でトークン情報を受け取り、ユーザ認証
// トークンからユーザIDを取り出し、dbからそのユーザIDのユーザの名前情報を取り出し、返す
func getUser(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserId(w, r)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	db, err := GetConnection()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	db_sql, err := db.DB()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer db_sql.Close()
	var user User
	// SELECT * FROM `users` WHERE user_id = '95daec2b-287c-4358-ba6f-5c29e1c3cbdf'
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
		return nil, err
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
func getUserId(w http.ResponseWriter, r *http.Request) (string, error) {
	tokenString := r.Header.Get("x-token")
	token, err := verifyToken(tokenString)
	if err != nil {
		return "", err
	}
	claims := token.Claims.(jwt.MapClaims)
	// claims = map[exp:1.629639808e+09 userId:bdd4056a-f113-424c-9951-1eaaaf853e5c]
	userId := claims["userId"].(string)
	return userId, nil
}

// -H "x-token:yyy"でトークン情報を受け取り、ユーザ認証
// -d {"name":"bbb"}で更新する名前情報を受け取る
// トークンからユーザIDを取り出し、dbからそのユーザIDのユーザの名前情報を更新
func updateUser(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserId(w, r)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	db, err := GetConnection()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	db_sql, err := db.DB()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer db_sql.Close()
	// dbでnameを更新
	// UPDATE `users` SET `name`='bbb' WHERE user_id = '95daec2b-287c-4358-ba6f-5c29e1c3cbdf'
	db.Model(&user).Where("user_id = ?", userId).Update("name", user.Name)
	Success(w, http.StatusOK, nil)
}

type Drawing struct {
	GachaID int `json:"gacha_id"`
	Times   int `json:"times"`
}

type Character struct {
	GachaCharacterID string `json:"gacha_character_id"`
	CharacterName    string `json:"character_name"`
	Weight           uint   `json:"weight"`
}

type UserCharacter struct {
	UserCharacterID  string `json:"user_character_id"`
	UserID           string `json:"user_id"`
	GachaCharacterID string `json:"gacha_character_id"`
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

// dbのgacha_charactersテーブルからgacha_id一覧を取得
// 引数のgachaIdがその中に含まれていたらtrue、含まれていなかったらfalseを返す
func gachaIdContains(gachaId int) (bool, error) {
	db, err := GetConnection()
	if err != nil {
		return false, err
	}
	db_sql, err := db.DB()
	if err != nil {
		return false, err
	}
	defer db_sql.Close()
	var gachaIds []int
	// SELECT gacha_id FROM `gacha_characters`
	db.Table("gacha_characters").Select("gacha_id").Scan(&gachaIds)
	for _, v := range gachaIds {
		if v == gachaId {
			return true, nil
		}
	}
	return false, nil
}

// localhost:8080/gacha/drawでガチャを引いて、キャラクターを取得
// -H "x-token:yyy"でトークン情報を受け取り、認証
// -d {"gacha_id":n, "times":x}でどのガチャを引くか、ガチャを何回引くかの情報を受け取る
func drawGacha(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserId(w, r)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var drawing Drawing
	if err := json.Unmarshal(body, &drawing); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	contains, err := gachaIdContains(drawing.GachaID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !contains {
		ErrorResponse(w, http.StatusBadRequest, "gacha_id is error.")
		return
	}
	if drawing.Times <= 0 {
		ErrorResponse(w, http.StatusBadRequest, "times is error.")
		return
	}

	db, err := GetConnection()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	db_sql, err := db.DB()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer db_sql.Close()
	charactersList, err := getCharacters(drawing.GachaID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	gachaCharacterIdsDrawed := drawGachaCharacterIds(charactersList, drawing.Times)
	var characterInfo CharacterResponse
	var results []CharacterResponse
	var userCharacters []UserCharacter
	count := 0
	for _, gacha_character_id := range gachaCharacterIdsDrawed {
		character := getCharacterInfo(charactersList, gacha_character_id)
		characterInfo = CharacterResponse{CharacterID: gacha_character_id, Name: character.CharacterName}
		results = append(results, characterInfo)
		userCharacterId, err := createUUId()
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		userCharacter := UserCharacter{UserCharacterID: userCharacterId, UserID: userId, GachaCharacterID: gacha_character_id}
		userCharacters = append(userCharacters, userCharacter)
		count += 1
		if count == 10000 {
			/*
				INSERT INTO `user_characters` (`user_character_id`,`user_id`,`gacha_character_id`)
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
			INSERT INTO `user_characters` (`user_character_id`,`user_id`,`gacha_character_id`)
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

// charactersListからキャラクターのgacha_character_idとweightを取り出しchoicesに格納
// times回分だけchoicesからWeighted Random Selectionを実行
func drawGachaCharacterIds(charactersList []Character, times int) []string {
	var choices []wr.Choice
	for i := 0; i < len(charactersList); i++ {
		choices = append(choices, wr.Choice{Item: charactersList[i].GachaCharacterID, Weight: charactersList[i].Weight})
	}
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
	//rand.Seed(time.Now().UTC().UnixNano())
	var gachaCharacterIdsDrawed []string
	for i := 0; i < times; i++ {
		chooser, _ := wr.NewChooser(choices...)
		gachaCharacterIdsDrawed = append(gachaCharacterIdsDrawed, chooser.Pick().(string))
	}
	return gachaCharacterIdsDrawed
}

// dbからキャラクターのgacha_character_id、名前、weightの情報を取得
// ガチャidが引数gacha_idのキャラクターに限る
func getCharacters(gacha_id int) ([]Character, error) {
	db, err := GetConnection()
	if err != nil {
		return nil, err
	}
	db_sql, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer db_sql.Close()
	var charactersList []Character
	/*
		SELECT gacha_characters.gacha_character_id, characters.character_name, rarities.weight
		FROM `gacha_characters`
		join characters
		on gacha_characters.character_id = characters.id
		join rarities
		on gacha_characters.rarity_id = rarities.id
		WHERE gacha_id = 1
	*/
	db.Table("gacha_characters").Select("gacha_characters.gacha_character_id, characters.character_name, rarities.weight").
		Joins("join characters on gacha_characters.character_id = characters.id").
		Joins("join rarities on gacha_characters.rarity_id = rarities.id").
		Where("gacha_id = ?", gacha_id).Scan(&charactersList)
	return charactersList, nil
}

// dbから全てのキャラクターのgacha_character_id、名前、weightの情報を取得
func getAllCharacters() ([]Character, error) {
	db, err := GetConnection()
	if err != nil {
		return nil, err
	}
	db_sql, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer db_sql.Close()
	var charactersList []Character
	/*
		SELECT gacha_characters.gacha_character_id, characters.character_name, rarities.weight
		FROM `gacha_characters`
		join characters
		on gacha_characters.character_id = characters.id
		join rarities
		on gacha_characters.rarity_id = rarities.id
	*/
	db.Table("gacha_characters").Select("gacha_characters.gacha_character_id, characters.character_name, rarities.weight").
		Joins("join characters on gacha_characters.character_id = characters.id").
		Joins("join rarities on gacha_characters.rarity_id = rarities.id").
		Scan(&charactersList)
	return charactersList, nil
}

// 引数のcharactersListからGachaCharacterIDが引数gacha_character_idのデータを取得
func getCharacterInfo(charactersList []Character, gacha_character_id string) Character {
	for i := 0; i < len(charactersList); i++ {
		if charactersList[i].GachaCharacterID == gacha_character_id {
			return charactersList[i]
		}
	}
	return Character{}
}

// dbのuser_charactersテーブルからuser_idが引数user_idのデータを取得
func getUserCharacters(user_id string) ([]UserCharacter, error) {
	db, err := GetConnection()
	if err != nil {
		return nil, err
	}
	db_sql, err := db.DB()
	if err != nil {
		return nil, err
	}
	defer db_sql.Close()
	var userCharactersList []UserCharacter
	// SELECT * FROM `user_characters`  WHERE (user_id = '703a0b0a-1541-487e-be5b-906e9541b021')
	db.Where("user_id = ?", user_id).Find(&userCharactersList)
	return userCharactersList, nil
}

// localhost:8080/character/listでユーザが所持しているキャラクター一覧情報を取得
// -H "x-token:yyy"でトークン情報を受け取り、認証
func getCharacterList(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserId(w, r)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	allCharactersList, err := getAllCharacters()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	userCharactersList, err := getUserCharacters(userId)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var characters []UserCharacterResponse
	var userCharacterInfo UserCharacterResponse
	if len(userCharactersList) == 0 {
		characters = make([]UserCharacterResponse, 0)
	} else {
		for _, v := range userCharactersList {
			gacha_character_id := v.GachaCharacterID
			character := getCharacterInfo(allCharactersList, gacha_character_id)
			characterName := character.CharacterName
			userCharacterInfo = UserCharacterResponse{UserCharacterID: v.UserCharacterID, CharacterID: gacha_character_id, Name: characterName}
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
