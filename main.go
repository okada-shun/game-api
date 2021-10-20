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
	gmfunction "local.packages/gmfunction"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Router作成
	router := mux.NewRouter()
	// URLと処理
	router.HandleFunc("/", home)
	// ユーザ関連API
	router.HandleFunc("/user/create", gmfunction.CreateUser).Methods("POST")
	router.HandleFunc("/user/get", gmfunction.GetUser).Methods("GET")
	router.HandleFunc("/user/update", gmfunction.UpdateUser).Methods("PUT")
	// ガチャ関連API
	router.HandleFunc("/gacha/draw", gmfunction.DrawGacha).Methods("POST")
	// キャラクター関連API
	router.HandleFunc("/character/list", gmfunction.GetCharacterList).Methods("GET")
	// 乱数のシード値を設定
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
	// ポートを8080で指定してRouter起動
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Hello Worldをlocalhost:8080画面に表示（後に削除予定）
func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}
