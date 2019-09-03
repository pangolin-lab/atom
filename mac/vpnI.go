package main

import "C"
import (
	"fmt"
	"github.com/pangolin-lab/atom/pipeProxy"
	"github.com/pangolin-lab/atom/wallet"
)

var proxyConf *pipeProxy.ProxyConfig = nil
var curProxy *pipeProxy.PipeProxy = nil

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
