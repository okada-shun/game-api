package api

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

// 16進数の秘密鍵文字列をイーサリアムアドレスに変換
func convertKeyToAddress(hexkey string) (common.Address, error) {
	// 16進数の秘密鍵文字列を読み込む
	privateKey, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		return common.Address{}, err
	}
	// 秘密鍵から公開鍵を生成
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	// 公開鍵からアドレスを生成
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address, nil
}

// コントラクトから、引数valだけゲームトークンを鋳造する
// 鋳造されたゲームトークンは、引数hexkeyの秘密鍵から生成されるアドレスに付与される
// トランザクションの送信者は、minter_private_key.txtの秘密鍵から生成されるアドレスである
func (c *Config) MintGmtoken(val int, hexkey string) error {
	// 16進数の秘密鍵文字列をアドレスに変換
	address, err := convertKeyToAddress(hexkey)
	if err != nil {
		return err
	}
	// トランザクションを送るアドレスの秘密鍵を読み込む
	privateKeyBytes, err := ioutil.ReadFile(c.MinterPrivateKey)
	if err != nil {
		return err
	}
	privateKey, err := crypto.HexToECDSA(string(privateKeyBytes))
	if err != nil {
		return err
	}
	// 秘密鍵からアドレスを生成
	fromAddress, err := convertKeyToAddress(string(privateKeyBytes))
	if err != nil {
		return err
	}
	// ナンスを生成
	nonce, err := c.Ethclient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}
	// 転送するイーサの量を設定(ここでは0)
	value := big.NewInt(0)
	// ガス価格を設定（SuggestGasPriceで平均のガス価格を取得）
	gasPrice, err := c.Ethclient.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	// fmt.Println(gasPrice) // 20000000000
	// GameTokenコントラクトのアドレスを読み込む
	contractAddressBytes, err := ioutil.ReadFile(c.ContractAddress)
	if err != nil {
		return err
	}
	contractAddress := common.HexToAddress(string(contractAddressBytes))
	// スマートコントラクトのmint関数
	mintFnSignature := []byte("mint(address,uint256)")
	// Keccak256ハッシュを生成
	hash := sha3.NewLegacyKeccak256()
	hash.Write(mintFnSignature)
	methodID := hash.Sum(nil)[:4]
	// fmt.Println(hexutil.Encode(methodID)) // 0x40c10f19
	paddedAddress := common.LeftPadBytes(address.Bytes(), 32)
	// fmt.Println(hexutil.Encode(paddedAddress)) // 0x00000000000000000000000001cf71979d4a11b0f15ca36b315cf9e77561c070
	// 転送するゲームトークンの量を設定
	amount := new(big.Int)
	amount.SetString(strconv.Itoa(val), 10)
	// 32バイト幅まで左側を0で埋める
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	// fmt.Println(hexutil.Encode(paddedAmount)) // 0x0000000000000000000000000000000000000000000000000000000000000064
	// メソッドIDと32バイト幅トークン量を連結
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	/*
		// EstimateGasでガス制限を推定
		gasLimit, err := c.Ethclient.EstimateGas(context.Background(), ethereum.CallMsg{
			To:   &contractAddress,
			Data: data,
		})
		if err != nil {
			return err
		}
	*/
	var gasLimit uint64 = 10000000
	// ナンス、コントラクトのアドレス、転送するイーサ量(0)、ガス制限、ガス価格、データからトランザクションを作成
	tx := types.NewTransaction(nonce, contractAddress, value, gasLimit, gasPrice, data)
	// チェーンID(ネットワークID)を取得
	chainID, err := c.Ethclient.NetworkID(context.Background())
	if err != nil {
		return err
	}
	// fmt.Println(chainID) // 5777
	// 送信者の秘密鍵を使用してトランザクションに署名
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return err
	}
	// fmt.Println(signedTx) // &{0xc00006e5a0 {13858186735851446500 49580235901 0xef4fa0} {<nil>} {<nil>} {<nil>}}
	// トランザクションを送信
	err = c.Ethclient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}

	// fmt.Printf("tx sent: %s", signedTx.Hash().Hex()) // tx sent: 0x8369c729025e98fd73e01c6e99724bb397bc58274b963b6eab75f1bd10dc39a1
	return nil
}

// コントラクトから、引数valだけゲームトークンを焼却する
// 引数hexkeyの秘密鍵から生成されるアドレスの持つゲームトークンを焼却する
// トランザクションの送信者は、引数hexkeyの秘密鍵から生成されるアドレスである
func (c *Config) BurnGmtoken(val int, hexkey string) error {
	// 16進数の秘密鍵文字列を読み込む
	privateKey, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		return err
	}
	// 16進数の秘密鍵文字列をアドレスに変換
	address, err := convertKeyToAddress(hexkey)
	if err != nil {
		return err
	}
	// トランザクションを送るアドレス
	fromAddress := address
	// ナンスを生成
	nonce, err := c.Ethclient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}
	// 転送するイーサの量を設定(ここでは0)
	value := big.NewInt(0)
	// ガス価格を設定（SuggestGasPriceで平均のガス価格を取得）
	gasPrice, err := c.Ethclient.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	// fmt.Println(gasPrice) // 20000000000
	// GameTokenコントラクトのアドレスを読み込む
	contractAddressBytes, err := ioutil.ReadFile(c.ContractAddress)
	if err != nil {
		return err
	}
	contractAddress := common.HexToAddress(string(contractAddressBytes))
	// スマートコントラクトのburn関数
	burnFnSignature := []byte("burn(uint256)")
	// Keccak256ハッシュを生成
	hash := sha3.NewLegacyKeccak256()
	hash.Write(burnFnSignature)
	methodID := hash.Sum(nil)[:4]
	// fmt.Println(hexutil.Encode(methodID)) // 0x42966c68
	// 転送するゲームトークンの量を設定
	amount := new(big.Int)
	amount.SetString(strconv.Itoa(val), 10)
	// 32バイト幅まで左側を0で埋める
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	// fmt.Println(hexutil.Encode(paddedAmount)) // 0x000000000000000000000000000000000000000000000000000000000000000a
	// メソッドIDと32バイト幅トークン量を連結
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAmount...)
	/*
		// EstimateGasでガス制限を推定
		gasLimit, err := c.Ethclient.EstimateGas(context.Background(), ethereum.CallMsg{
			To:   &contractAddress,
			Data: data,
		})
		if err != nil {
			return err
		}
	*/
	var gasLimit uint64 = 10000000
	// ナンス、コントラクトのアドレス、転送するイーサ量(0)、ガス制限、ガス価格、データからトランザクションを作成
	tx := types.NewTransaction(nonce, contractAddress, value, gasLimit, gasPrice, data)
	// チェーンID(ネットワークID)を取得
	chainID, err := c.Ethclient.NetworkID(context.Background())
	if err != nil {
		return err
	}
	// fmt.Println(chainID) // 5777
	// 送信者の秘密鍵を使用してトランザクションに署名
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return err
	}
	// fmt.Println(signedTx) // &{0xc00012bf20 {13858187368593859236 638888717301 0xef4fa0} {<nil>} {<nil>} {<nil>}}
	// トランザクションを送信
	err = c.Ethclient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}

	// fmt.Printf("tx sent: %s", signedTx.Hash().Hex()) // tx sent: 0xf98c12a353eceacafe606397493d0d321628f1a70bb147697d1539a2a9ca9199
	return nil
}
