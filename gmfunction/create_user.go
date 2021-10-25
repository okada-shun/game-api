package gmfunction

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	transaction "local.packages/transaction"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	UserID     string `json:"user_id"`
	Name       string `json:"name"`
	PrivateKey string `json:"private_key"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

// localhost:8080/user/createでユーザ情報を作成
// -d {"name":"aaa"}で名前データを受け取る
// UUIDでユーザIDを生成する
// ユーザIDからjwtでトークンを作成し、トークンを返す
func CreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	userId, err := createUUId()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user.UserID = userId
	// 新規ユーザの秘密鍵を生成
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)[2:]
	user.PrivateKey = privateKeyHex
	// ゲームトークンを100だけ鋳造し、新規ユーザに付与
	if err := transaction.MintGmtoken(100, user.PrivateKey); err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
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
	/*
		INSERT INTO `users` (`user_id`,`name`,`private_key`)
		VALUES ('95daec2b-287c-4358-ba6f-5c29e1c3cbdf','aaa','6e7eada90afb7e84bf5b4498c6adaa2d4014904644637d5fb355266944fbf93a')
	*/
	db.Create(&user)
	// ユーザIDの文字列からjwtでトークン作成
	token, err := createToken(userId)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// token = "生成されたトークンの文字列"
	RespondWithJSON(w, http.StatusOK, &TokenResponse{
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

// UUIDを生成
func createUUId() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	uu := u.String()
	return uu, nil
}
