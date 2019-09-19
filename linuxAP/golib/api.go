package golib

import (
	"github.com/pangolink/proton-node/account"
	"github.com/btcsuite/btcutil/base58"
	"github.com/proton-lab/autom/ethereum"
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

//Create Eth Account
func LibCreateEthAccount(saveDir,password string) string  {
	return ethereum.CreateEthAccount2(password,saveDir)
}

//Import Eth Account
func LibImportEthAccount(accFile,saveDir,password string) string  {
	return ethereum.ImportEthAccount(accFile,saveDir,password)
}

//VerifyEthAccount
func LibVerifyEthAccount(ciperTxt,passphrase string) bool {
	return ethereum.VerifyEthAccount(ciperTxt,passphrase)
}

//Load Eth Account By Proton Account
func LibLoadEthAcctByProtonAcct(protonAcct string) string  {
	return ethereum.BoundEth(protonAcct)
}

//Test and Get Balance of the Eth account
func LibGetEthAcctBalance(ethAcct string) (float64,int)  {

	ethB,no:=ethereum.BasicBalance(ethAcct)
	if ethB == nil{
		return  0,0
	}

	return ethereum.ConvertByDecimal(ethB),no
}

//Bind Proton Account
func LibBindProtonAccount(protonAcct,cipherTxt,password string) (string,error)  {
	return ethereum.Bind(protonAcct,cipherTxt,password)
}

//UnBind Proton Account
func LibUnbindProtonAccount(protonAcct,cipherTxt,password string) (string,error)  {
	return ethereum.Unbind(protonAcct,cipherTxt,password)
}

