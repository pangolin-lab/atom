package golib


import (
	"github.com/pangolink/proton-node/account"
	"github.com/btcsuite/btcutil/base58"
)

//Create Proton Account
func LibCreateAccount(password string) (address,cipherTxt string) {

	key, err := account.GenerateKey(password)
	if err != nil {
		return
	}
	address = key.ToNodeId().String()
	cipherTxt = base58.Encode(key.LockedKey)

	return
}


