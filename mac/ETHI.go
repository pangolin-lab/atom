package main

import "C"

import (
	"encoding/json"
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
func WalletBalance(address string) (*C.char, *C.char) {
	eth, token := ethereum.TokenBalance(address)
	return C.CString(fmt.Sprintf("%.8f", eth)), C.CString(fmt.Sprintf("%.8f", token))
}

//export WalletVerify
func WalletVerify(cipher, auth string) bool {
	return account.VerifyWallet(([]byte)(cipher), auth)
}

//export MinerPoolAddresses
func MinerPoolAddresses() *C.char {

	arr := ethereum.PoolAddressList()
	addr := make([]string, len(arr))

	for i := 0; i < len(arr); i++ {
		addr = append(addr, arr[i].Hex())
	}

	buf, _ := json.Marshal(addr)
	return C.CString(string(buf))
}

//export MinerDetails
func MinerDetails(addr string) *C.char {
	return C.CString(ethereum.PoolDetails(addr))
}

//export MinerPoolList
func MinerPoolList() *C.char {
	jsonStr := ethereum.PoolListWithDetails()
	return C.CString(jsonStr)
}

//export MySubPools
func MySubPools(addr string) *C.char {
	jsonStr := ethereum.MySubPools(addr)
	return C.CString(jsonStr)
}

//export MySubPoolsWithDetails
func MySubPoolsWithDetails(addr string) *C.char {
	jsonStr := ethereum.MySubPoolsWithDetails(addr)
	return C.CString(jsonStr)
}
