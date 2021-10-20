package gmfunction

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
)

// -H "x-token:yyy"でトークン情報を受け取り、ユーザ認証
// -d {"name":"bbb"}で更新する名前データを受け取る
// トークンからユーザIDを取り出し、dbからそのユーザIDのユーザの情報を更新
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserId(w, r)
	if err != nil {
		ReplyResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		ReplyResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		ReplyResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	db, err := GetConnection()
	if err != nil {
		ReplyResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	db_sql, err := db.DB()
	if err != nil {
		ReplyResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	defer db_sql.Close()
	// dbでnameとaddressを更新
	// UPDATE `users` SET `name`='bbb' WHERE user_id = '95daec2b-287c-4358-ba6f-5c29e1c3cbdf'
	db.Model(&user).Where("user_id = ?", userId).Update("name", user.Name)
	ReplyResponse(w, http.StatusOK, "", nil)
}
