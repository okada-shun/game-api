package main

import (
	crand "crypto/rand"
	"fmt"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"github.com/gorilla/mux"
	api "local.packages/api"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// configインスタンスを作成
	config := api.NewConfig()
	// DBコネクションを閉じる
	db_sql, err := config.DB.DB()
	if err != nil {
		fmt.Println(err)
	}
	defer db_sql.Close()
	// 乱数のシード値を設定
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
	// サーバー起動
	startServer(config)
}

// サーバー起動
func startServer(config *api.Config) {
	// Router作成
	router := mux.NewRouter()
	// URLと処理
	router.HandleFunc("/", home)
	// ユーザ関連API
	router.HandleFunc("/user/create", config.CreateUser).Methods("POST")
	router.HandleFunc("/user/get", config.GetUser).Methods("GET")
	router.HandleFunc("/user/update", config.UpdateUser).Methods("PUT")
	// ガチャ関連API
	router.HandleFunc("/gacha/draw", config.DrawGacha).Methods("POST")
	// キャラクター関連API
	router.HandleFunc("/character/list", config.GetCharacterList).Methods("GET")
	// ポートを8080で指定してRouter起動
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Hello Worldをlocalhost:8080画面に表示
func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}
