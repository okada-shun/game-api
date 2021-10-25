package gmfunction

import (
	"net/http"
	_ "github.com/go-sql-driver/mysql"
)

type UserCharacterResponse struct {
	UserCharacterID string `json:"userCharacterID"`
	CharacterID     string `json:"characterID"`
	Name            string `json:"name"`
}

// getCharacterList関数で返される
type CharactersResponse struct {
	Characters []UserCharacterResponse `json:"characters"`
}

// localhost:8080/character/listでユーザが所持しているキャラクター一覧情報を取得
// -H "x-token:yyy"でトークン情報を受け取り、認証
func GetCharacterList(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserId(w, r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	allCharactersList, err := getAllCharacters()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	userCharactersList, err := getUserCharacters(userId)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
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
	RespondWithJSON(w, http.StatusOK, &CharactersResponse{
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
