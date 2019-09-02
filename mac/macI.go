package main

import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/pipeProxy"
	"github.com/pangolin-lab/atom/wallet"
	"github.com/pangolink/miner-pool/account"
)

var proxyConf *pipeProxy.ProxyConfig = nil
var curProxy *pipeProxy.PipeProxy = nil

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

//export LibIsInit
func LibIsInit() bool {
	return curProxy != nil
}

//export LibInitProxy
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

//export LibCreateProxy
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
//export LibProxyRun
func LibProxyRun() {
	if curProxy == nil {
		return
	}
	fmt.Println("start proxy success.....")

	curProxy.Proxying()
	curProxy.Finish()
	curProxy = nil
}

//export LibStopClient
func LibStopClient() {
	if curProxy == nil {
		return
	}
	curProxy.Finish()
	return
}
