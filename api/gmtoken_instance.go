package api

import (
	"io/ioutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gmtoken "local.packages/gmtoken"
)

// イーサリアムネットワークに接続するクライアントを取得
func getEthclient(url string) (*ethclient.Client, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Gmtokenのインスタンスを取得
func getGmtokenInstance(url string, addressfile string) (*gmtoken.Gmtoken, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	// GameTokenコントラクトのアドレスを読み込む
	contractAddressBytes, err := ioutil.ReadFile(addressfile)
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