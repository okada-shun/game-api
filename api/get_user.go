package api

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/go-sql-driver/mysql"
)

// getUser関数で返されるユーザの名前、アドレス、ゲームトークン残高の情報が入る
type UserResponse struct {
	Name           string `json:"name"`
	Address        string `json:"address"`
	GmtokenBalance int    `json:"gmtoken_balance"`
}

// -H "x-token:yyy"でトークン情報を受け取り、ユーザ認証
// トークンからユーザIDを取り出し、dbからそのユーザIDのユーザの名前と秘密鍵データを取り出す
// 秘密鍵からユーザアドレスを生成
// コントラクトからそのユーザアドレスのゲームトークン残高を取り出し、返す
func (c *Config) GetUser(w http.ResponseWriter, r *http.Request) {
	userId, err := c.getUserId(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	var user User
	// SELECT * FROM `users` WHERE user_id = '95daec2b-287c-4358-ba6f-5c29e1c3cbdf'
	c.DB.Where("user_id = ?", userId).Find(&user)
	address, balance, err := c.getAddressBalance(user.PrivateKey)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, &UserResponse{
		Name:           user.Name,
		Address:        address.String(),
		GmtokenBalance: balance,
	})
	// {"name":"aaa","address":"0x7a242084216fC7810aAe02c6deE5D9092C6B8fb9","gmtoken_balance":40}が返る
	// 有効期限が切れると{"code":400,"message":"Token is expired"}が返る
}

// jwtトークンを認証する
func (c *Config) verifyToken(tokenString string) (*jwt.Token, error) {
	// 秘密鍵を取得
	signBytes, err := ioutil.ReadFile(c.Idrsa)
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
// トークンからユーザID情報を取り出し、返す
func (c *Config) getUserId(r *http.Request) (string, error) {
	tokenString := r.Header.Get("x-token")
	token, err := c.verifyToken(tokenString)
	if err != nil {
		return "", err
	}
	claims := token.Claims.(jwt.MapClaims)
	// claims = map[exp:1.629639808e+09 userId:bdd4056a-f113-424c-9951-1eaaaf853e5c]
	userId := claims["userId"].(string)
	return userId, nil
}

// 引数の秘密鍵hexkeyからアドレスを生成
// コントラクトからそのアドレスのゲームトークン残高を取り出す
// アドレスと残高を返す
func (c *Config) getAddressBalance(hexkey string) (common.Address, int, error) {
	address, err := convertKeyToAddress(hexkey)
	if err != nil {
		return common.Address{}, 0, err
	}
	bal, err := c.GmtokenInstance.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		return common.Address{}, 0, err
	}
	balance, _ := strconv.Atoi(bal.String())
	return address, balance, nil
}