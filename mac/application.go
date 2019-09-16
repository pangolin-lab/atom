package main

import "C"
import (
	"fmt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/payment"
	"github.com/pangolin-lab/atom/proxy"
	"github.com/pangolin-lab/atom/utils"
	"path/filepath"
)

const (
	Success = iota
	ErrOpenWallet
	ErrCreateDir
	ErrInitProtocol
	ErrInitDataCache
	ErrInitVpnService
	ErrVpnServiceExit

	WalletFile      = "wallet.json"
	ReceiptDataBase = "accountant"
	BlockDataBase   = "blockData"
)

type MacApp struct {
	proxy   *proxy.VpnProxy
	ppp     payment.PacketPaymentProtocol
	cache   *payment.BlockChainDataCache
	service *proxy.VpnProxy
	err     chan error
}

var _appInstance = &MacApp{
	err: make(chan error, 10),
}

func initEthereumConf(tokenAddr, payChanAddr, apiUrl string) {
	if tokenAddr != "" {
		var tmp = make([]byte, len(tokenAddr))
		copy(tmp, ([]byte)(tokenAddr))
		ethereum.Conf.Token = string(tmp)
	}
	if payChanAddr != "" {
		var tmp = make([]byte, len(payChanAddr))
		copy(tmp, ([]byte)(payChanAddr))
		ethereum.Conf.MicroPaySys = string(tmp)
	}
	if apiUrl != "" {
		var tmp = make([]byte, len(apiUrl))
		copy(tmp, ([]byte)(apiUrl))
		ethereum.Conf.EthApiUrl = string(tmp)
	}

	fmt.Println(ethereum.Conf.String())
}

//export initApp
func initApp(tokenAddr, payChanAddr, apiUrl, baseDir, srvAddr string) (int, *C.char) {
	initEthereumConf(tokenAddr, payChanAddr, apiUrl)

	if err := utils.TouchDir(baseDir); err != nil {
		return ErrCreateDir, C.CString(err.Error())
	}
	walletPath := filepath.Join(baseDir, string(filepath.Separator), WalletFile)
	receiptPath := filepath.Join(baseDir, string(filepath.Separator), ReceiptDataBase)

	protocol, err := payment.InitProtocol(walletPath, receiptPath)
	if err != nil {
		return ErrInitProtocol, C.CString(err.Error())
	}

	_appInstance.ppp = protocol
	cachePath := filepath.Join(baseDir, string(filepath.Separator), BlockDataBase)
	addr, _ := _appInstance.ppp.WalletAddr()

	cc, err := payment.InitBlockDataCache(cachePath, addr)
	if err != nil {
		return ErrInitDataCache, C.CString(err.Error())
	}
	_appInstance.cache = cc

	return Success, nil
}

//export startService
func startService(srvAddr string) (int, *C.char) {
	srv, err := proxy.NewProxyService(srvAddr, nil)
	if err != nil {
		return ErrInitVpnService, C.CString(err.Error())
	}
	_appInstance.service = srv

	go srv.Accepting(_appInstance.err, proxy.Socks5Target, _appInstance.ppp)
	ret := <-_appInstance.err

	stopService()
	return ErrVpnServiceExit, C.CString(ret.Error())
}

//export stopService
func stopService() {
	_appInstance.err <- fmt.Errorf("stopped by user")
}

//export openWallet
func openWallet(auth string) (int, *C.char) {

	if err := _appInstance.ppp.OpenPacketWallet(auth); err != nil {
		return ErrOpenWallet, C.CString(err.Error())
	}
	return Success, C.CString("")
}
