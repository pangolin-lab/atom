package main

import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolink/miner-pool/account"
)

//export WalletVerify
func WalletVerify(cipher, auth string) bool {
	return account.VerifyWallet(([]byte)(cipher), auth)
}

//export TransferEth
func TransferEth(auth, target string, sum float64) (*C.char, *C.char) {

	w, e := _appInstance.protocol.Wallet(auth)
	if e != nil {
		return nil, C.CString(e.Error())
	}
	tx, e := ethereum.TransferEth(target, sum, w.SignKey())
	if e != nil {
		return nil, C.CString(e.Error())
	}
	fmt.Printf("tx sent: %s", tx)
	return C.CString(tx), C.CString("")
}

//export TransferLinToken
func TransferLinToken(auth, target string, sum float64) (*C.char, *C.char) {

	w, err := _appInstance.protocol.Wallet(auth)
	if err != nil {
		return nil, C.CString(err.Error())
	}
	tx, e := ethereum.TransferLinToken(target, sum, w.SignKey())
	if e != nil {
		return nil, C.CString(e.Error())
	}

	fmt.Printf("tx sent: %s", tx)
	return C.CString(tx), nil
}

//export SyncWalletInfo
func SyncWalletInfo() {
	go _appInstance.protocol.SyncWalletBalance()
}

//export LoadWalletInfo
func LoadWalletInfo() *C.char {
	abi := _appInstance.protocol.AccBookInfo()
	if abi == nil {
		return nil
	}
	data, err := json.Marshal(abi)
	if err != nil {
		return nil
	}
	return C.CString(string(data))
}

//export NewWallet
func NewWallet(auth string) (bool, *C.char) {

	_, e := _appInstance.protocol.NewWallet(auth, _appInstance.conf.walletPath)
	if e != nil {
		return false, C.CString(e.Error())
	}

	if e := _appInstance.ReInitApp(); e != nil {
		return false, C.CString(e.Error())
	}

	return true, nil
}
