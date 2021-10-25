package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	wr "github.com/mroth/weightedrand"
	transaction "local.packages/transaction"
	_ "github.com/go-sql-driver/mysql"
)

type DrawingGacha struct {
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

// localhost:8080/gacha/drawでガチャを引いて、キャラクターを取得
// -H "x-token:yyy"でトークン情報を受け取り、認証
// -d {"gacha_id":n, "times":x}でどのガチャを引くか、ガチャを何回引くかの情報を受け取る
func DrawGacha(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserId(w, r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	var drawingGacha DrawingGacha
	if err := json.Unmarshal(body, &drawingGacha); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	contains, err := gachaIdContains(drawingGacha.GachaID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !contains {
		RespondWithError(w, http.StatusBadRequest, "gacha_id is error.")
		return
	}
	// 0以下回だけガチャを引くことは出来ない
	if drawingGacha.Times <= 0 {
		RespondWithError(w, http.StatusBadRequest, "times is error.")
		return
	}
	enoughBal, err := checkBalance(userId, drawingGacha.Times)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !enoughBal {
		RespondWithError(w, http.StatusBadRequest, "Balance of GameToken is not enough.")
		return
	}

	db, err := GetConnection()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	db_sql, err := db.DB()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer db_sql.Close()
	var user User
	// SELECT * FROM `users` WHERE user_id = '95daec2b-287c-4358-ba6f-5c29e1c3cbdf'
	db.Where("user_id = ?", userId).Find(&user)
	// drawingGacha.Times分だけゲームトークンを焼却
	if err := transaction.BurnGmtoken(drawingGacha.Times, user.PrivateKey); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	charactersList, err := getCharacters(drawingGacha.GachaID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	gachaCharacterIdsDrawed := drawGachaCharacterIds(charactersList, drawingGacha.Times)
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
			RespondWithError(w, http.StatusInternalServerError, err.Error())
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
	RespondWithJSON(w, http.StatusOK, &ResultResponse{
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

// dbのusersテーブルからuser_idが引数userIdのユーザ情報を取得
// コントラクトからそのユーザアドレスのゲームトークン残高を取得
// 引数のtimesが残高以下だったらtrue、残高より大きかったらfalseを返す
func checkBalance(userId string, times int) (bool, error) {
	db, err := GetConnection()
	if err != nil {
		return false, err
	}
	db_sql, err := db.DB()
	if err != nil {
		return false, err
	}
	defer db_sql.Close()
	var user User
	// SELECT * FROM `users` WHERE user_id = '95daec2b-287c-4358-ba6f-5c29e1c3cbdf'
	db.Where("user_id = ?", userId).Find(&user)
	
	_, balance, err := getAddressBalance(user.PrivateKey)
	if err != nil {
		return false, err
	}
	
	return times <= balance, nil
}

// charactersListからキャラクターのgacha_character_idとweightを取り出しchoicesに格納
// times回分だけchoicesからWeighted Random Selectionを実行
func drawGachaCharacterIds(charactersList []Character, times int) []string {
	var choices []wr.Choice
	for i := 0; i < len(charactersList); i++ {
		choices = append(choices, wr.Choice{Item: charactersList[i].GachaCharacterID, Weight: charactersList[i].Weight})
	}
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

// 引数のcharactersListからGachaCharacterIDが引数gacha_character_idのデータを取得
func getCharacterInfo(charactersList []Character, gacha_character_id string) Character {
	for i := 0; i < len(charactersList); i++ {
		if charactersList[i].GachaCharacterID == gacha_character_id {
			return charactersList[i]
		}
	}
	return Character{}
}
