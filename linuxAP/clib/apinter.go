package main

import "C"
import (
	"fmt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/pipeProxy"
	"github.com/pangolin-lab/atom/wallet"
	"github.com/pangolink/proton-node/account"
	"github.com/pangolin-lab/atom/linuxAP/golib"
)

var proxyConf *pipeProxy.ProxyConfig = nil
var curProxy *pipeProxy.PipeProxy = nil

//Create Proton Account
func LibCreateAccount(password string) (*C.char, *C.char) {

	addr,cipherTxt:=golib.LibCreateAccount(password)

	return C.CString(addr), C.CString(cipherTxt)
}

//Create Ethereum Account
func LibCreateEthAccount(password, directory string) *C.char {
	return C.CString(ethereum.CreateEthAccount(password, directory))
}

//Test have proxy exists?
func LibIsInit() bool {
	return curProxy != nil
}

//Test Account is valid
func LibVerifyAccount(cipherTxt, address, password string) bool {
	if _, err := account.AccFromString(address, cipherTxt, password); err != nil {
		return false
	}
	return true
}

//Test proton address is valid
func LibIsProtonAddress(address string) bool {
	return account.ID(address).IsValid()
}

//Init proxy
func LibInitProxy(addr, cipher, url, boot, path string) bool {
	proxyConf = &pipeProxy.ProxyConfig{
		WConfig: &wallet.WConfig{
			BCAddr:     addr,
			Cipher:     cipher,
			SettingUrl: url,
			Saver:      nil,
		},
		BootNodes: boot,
	}

	mis := proxyConf.FindBootServers(path)
	if len(mis) == 0 {
		fmt.Println("no valid boot strap node")
		return false
	}

	proxyConf.ServerId = mis[0]
	return true
}

//Create proxy
func LibCreateProxy(password, locSer string) bool {

	if proxyConf == nil {
		fmt.Println("init the proxy configuration first please")
		return false
	}

	if curProxy != nil {
		fmt.Println("stop the old instance first please")
		return true
	}

	fmt.Println(proxyConf.String())

	w, err := wallet.NewWallet(proxyConf.WConfig, password)
	if err != nil {
		fmt.Println(err)
		return false
	}

	proxy, e := pipeProxy.NewProxy(locSer, w, NewTunReader())
	if e != nil {
		fmt.Println(e)
		return false
	}
	curProxy = proxy
	return true
}

//TODO:: inner error call back
//Run proxy
func LibProxyRun() {
	if curProxy == nil {
		return
	}
	fmt.Println("start proxy success.....")

	curProxy.Proxying()
	curProxy.Finish()
	curProxy = nil
}

//stop client
func LibStopClient() {
	if curProxy == nil {
		return
	}
	curProxy.Finish()
	return
}

//export LibLoadEthAddrByProtonAddr
func LibLoadEthAddrByProtonAddr(protonAddr string) *C.char {
	return C.CString(ethereum.BoundEth(protonAddr))
}

//export LibEthBindings
func LibEthBindings(ETHAddr string) (float64, int) {
	ethB, no := ethereum.BasicBalance(ETHAddr)
	if ethB == nil {
		return 0, 0
	}
	return ethereum.ConvertByDecimal(ethB), no
}

//export LibImportEthAccount
func LibImportEthAccount(file, dir, pwd string) *C.char {
	addr := ethereum.ImportEthAccount(file, dir, pwd)
	return C.CString(addr)
}

//export LibBindProtonAddr
func LibBindProtonAddr(protonAddr, cipherKey, password string) (*C.char, *C.char) {

	tx, err := ethereum.Bind(protonAddr, cipherKey, password)
	if err != nil {
		fmt.Printf("\nBind proton addr(%s) err:%s", protonAddr, err)
		return C.CString(""), C.CString(err.Error())
	}

	return C.CString(tx), C.CString("")
}

//export LibUnbindProtonAddr
func LibUnbindProtonAddr(protonAddr, cipherKey, password string) (*C.char, *C.char) {

	tx, err := ethereum.Unbind(protonAddr, cipherKey, password)
	if err != nil {
		fmt.Printf("\nBind proton addr(%s) err:%s", protonAddr, err)
		return C.CString(""), C.CString(err.Error())
	}

	return C.CString(tx), C.CString("")
}
