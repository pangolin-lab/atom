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
func TransferEth(cipher, auth, target string, sum float64) (*C.char, *C.char) {

	w, e := account.DecryptWallet([]byte(cipher), auth)
	if e != nil {
		return C.CString(""), C.CString(e.Error())
	}
	tx, e := ethereum.TransferEth(target, sum, w.SignKey())
	if e != nil {
		return C.CString(""), C.CString(e.Error())
	}
	fmt.Printf("tx sent: %s", tx)
	return C.CString(tx), C.CString("")
}

//export TransferLinToken
func TransferLinToken(cipher, auth, target string, sum float64) (*C.char, *C.char) {

	w, e := account.DecryptWallet([]byte(cipher), auth)
	if e != nil {
		return C.CString(""), C.CString(e.Error())
	}

	tx, e := ethereum.TransferLinToken(target, sum, w.SignKey())
	if e != nil {
		return C.CString(""), C.CString(e.Error())
	}

	fmt.Printf("tx sent: %s", tx)
	return C.CString(tx), C.CString("")
}

//export WalletBalance
func WalletBalance() *C.char {
	wi := _appInstance.ppp.AccountBook()
	if wi == nil {
		return nil
	}
	data, err := json.Marshal(wi)
	if err != nil {
		_appInstance.DataSyncedFailed(err)
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
	_appInstance.ppp.Finish()

	if err := ioutil.WriteFile(_appInstance.conf.walletDir, wJson, 0644); err != nil {
		return false, C.CString(err.Error())
	}
	return true, nil
}
