package api

import (
	"fmt"
	"net/http"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
	gmtoken "local.packages/gmtoken"
)

// Idrsa: jwtトークンの作成・認証に使用するサーバーの秘密鍵
// MinterPrivateKey: MintGmtoken関数で使用する、Minterの秘密鍵
// ContractAddress: GameTokenコントラクトのアドレス
type Config struct {
	Idrsa string
	MinterPrivateKey string
	ContractAddress string
	GmtokenInstance *gmtoken.Gmtoken
	DB *gorm.DB
	Ethclient *ethclient.Client
}

// main関数内でconfigインスタンス作成
func NewConfig() *Config {
	return &Config{
		Idrsa: "../.ssh/id_rsa",
		MinterPrivateKey: "./minter_private_key.txt",
		ContractAddress: "./GameToken_address.txt",
		GmtokenInstance: newGmtokenInstance("ws://localhost:7545", "./GameToken_address.txt"),
		DB: newDBConnection("../.ssh/mysql_password", "../.ssh/mysql_user"),
		Ethclient: newEthclient("ws://localhost:7545"),
	}
}

// Gmtokenのインスタンスを返す
func newGmtokenInstance(url string, addressfile string) *gmtoken.Gmtoken {
	gmtokenInstance, err := getGmtokenInstance(url, addressfile)
	if err != nil {
		fmt.Println(err.Error(), http.StatusInternalServerError)
	}
	return gmtokenInstance
}

// DBコネクションを返す
func newDBConnection(passwordfile string, userfile string) *gorm.DB {
	db, err := GetConnection(passwordfile, userfile)
	if err != nil {
		fmt.Println(err.Error(), http.StatusInternalServerError)
	}
	return db
}

// イーサリアムネットワークに接続するクライアントを返す
func newEthclient(url string) *ethclient.Client {
	client, err := getEthclient(url)
	if err != nil {
		fmt.Println(err.Error(), http.StatusInternalServerError)
	}
	return client
}
