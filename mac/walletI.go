package main

import "C"
import (
	"fmt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolink/miner-pool/account"
)

//export InitBlockChain
func InitBlockChain(tokenAddr, microPayAddr, apiUrl string) {
	if tokenAddr != "" {
		var tmp = make([]byte, len(tokenAddr))
		copy(tmp, ([]byte)(tokenAddr))
		ethereum.Conf.Token = string(tmp)
	}
	if microPayAddr != "" {
		var tmp = make([]byte, len(microPayAddr))
		copy(tmp, ([]byte)(microPayAddr))
		ethereum.Conf.MicroPaySys = string(tmp)
	}
	if apiUrl != "" {
		var tmp = make([]byte, len(apiUrl))
		copy(tmp, ([]byte)(apiUrl))
		ethereum.Conf.EthApiUrl = string(tmp)
	}

	fmt.Println(ethereum.Conf.String())
}

//export NewWallet
func NewWallet(password string) *C.char {
	w := account.NewWallet()
	if w == nil {
		fmt.Print("Create new Wallet failed")
		return nil
	}

	wJson, err := w.EncryptWallet(password)
	if err != nil {
		fmt.Print(err)
		return nil
	}
	return C.CString(string(wJson))
}

//export WalletBalance
func WalletBalance(address string) (float64, float64) {
	return ethereum.TokenBalance(address)
}

//export WalletVerify
func WalletVerify(cipher, auth string) bool {
	return account.VerifyWallet(([]byte)(cipher), auth)
}
