package gmfunction

import (
	"io/ioutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gmtoken "local.packages/gmtoken"
)

// Gmtokenのインスタンスを取得
func getGmtokenInstance() (*gmtoken.Gmtoken, error) {
	// localhost:7545(ganacheテストネット)に接続するクライアントを取得
	client, err := ethclient.Dial("ws://localhost:7545")
	if err != nil {
		return nil, err
	}
	// GameTokenコントラクトのアドレスを読み込む
	contractAddressBytes, err := ioutil.ReadFile("./GameToken_address.txt")
	if err != nil {
		return nil, err
	}
	contractAddress := common.HexToAddress(string(contractAddressBytes))
	// 上のコントラクトアドレスのGmtokenインスタンスを作成
	instance, err := gmtoken.NewGmtoken(contractAddress, client)
	if err != nil {
		return nil, err
	}
	return instance, nil
}