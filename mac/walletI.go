package main

import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolink/miner-pool/account"
	"io/ioutil"
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
func SyncWalletInfo() *C.char {
	wi := _appInstance.protocol.SyncWalletData()
	if wi == nil {
		return nil
	}
	data, err := json.Marshal(wi)
	if err != nil {
		return nil
	}
	return C.CString(string(data))
}

//export NewWallet
func NewWallet(auth string) (bool, *C.char) {

	w, err := account.NewWallet()
	if err != nil {
		return false, C.CString(err.Error())
	}

	wJson, err := w.EncryptWallet(auth)
	if err != nil {
		return false, C.CString(err.Error())
	}
	_appInstance.protocol.Finish()

	if err := ioutil.WriteFile(_appInstance.conf.walletPath, wJson, 0644); err != nil {
		return false, C.CString(err.Error())
	}
	return true, nil
}
