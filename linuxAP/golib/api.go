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

//Verify Proton Account
func LibVerifyProtonAddress(address string) bool {
	return account.ID(address).IsValid()
}


//Create Eth Account
func LibCreateEthAccount(password,dir string) string  {

	return ethereum.CreateEthAccount2(password,dir)
}

//Import Eth Account
func LibImportEthAccount(accFile,dir,password string) string  {
	return ethereum.ImportEthAccount(accFile,dir,password)
}



//VerifyEthAccount
func LibVerifyEthAccount(ciperTxt,passphrase string) bool {
	return ethereum.VerifyEthAccount(ciperTxt,passphrase)
}

//Load Eth Account By Proton Address
func LibLoadEthAcctByProtonAcct(protonAddr string) string  {
	return ethereum.BoundEth(protonAddr)
}

//Test and Get Balance of the Eth address
func LibGetEthAcctBalance(ethAddr string) (float64,int)  {

	ethB,no:=ethereum.BasicBalance(ethAddr)
	if ethB == nil{
		return  0,0
	}

	return ethereum.ConvertByDecimal(ethB),no

}

//Bind Proton Account
func LibBindProtonAccount(protonAddr,cipherTxt,password string) (string,error)  {
	return ethereum.Bind(protonAddr,cipherTxt,password)
}

//UnBind Proton Account
func LibUnbindProtonAccount(protonAddr,cipherTxt,password string) (string,error)  {
	return ethereum.Unbind(protonAddr,cipherTxt,password)
}

