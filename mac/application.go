package main

/*
#include "callback.h"
*/
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
	ErrCreateDir
	ErrInitProtocol
	ErrInitDataCache
	ErrInitVpnService
	ErrVpnServiceExit
	ErrOpenPayChannel
	ErrNoSuchPool

	WalletFile      = "wallet.json"
	ReceiptDataBase = "accountant"
	BlockDataBase   = "blockData"
)

type appConf struct {
	baseDir     string
	walletPath  string
	receiptPath string
	cachePath   string
}

func (ac appConf) String() string {

	str := fmt.Sprintf("\n++++++++++++++++++++++++++++++++++++++++++++++++++++"+
		"\n base dir:%s"+
		"\n wallet path:%s"+
		"\n receipt path:%s"+
		"\n++++++++++++++++++++++++++++++++++++++++++++++++++++",
		ac.baseDir, ac.walletPath, ac.receiptPath)

	return str
}

type MacApp struct {
	conf     appConf
	protocol payment.PacketPaymentProtocol
	dataSrv  *payment.BlockChainDataService
	service  *proxy.VpnProxy
	err      chan error
	sysImp   C.SystemActionCallBack
	dataImp  C.BlockChainDataSyncNotifier
}

var _appInstance = &MacApp{
	err: make(chan error, 10),
}

func initEthereumConf(tokenAddr, payChanAddr, apiUrl string) {
	if tokenAddr != "" {
		var tmp = make([]byte, len(tokenAddr))
		copy(tmp, tokenAddr)
		ethereum.Conf.Token = string(tmp)
	}
	if payChanAddr != "" {
		var tmp = make([]byte, len(payChanAddr))
		copy(tmp, payChanAddr)
		ethereum.Conf.MicroPaySys = string(tmp)
	}
	if apiUrl != "" {
		var tmp = make([]byte, len(apiUrl))
		copy(tmp, apiUrl)
		ethereum.Conf.EthApiUrl = string(tmp)
	}
	fmt.Println(ethereum.Conf.String())
	fmt.Println("init ethereum config success......")
}

//export initApp
func initApp(tokenAddr, payChanAddr, apiUrl, baseDir string,
	si C.SystemActionCallBack, di C.BlockChainDataSyncNotifier) (int, *C.char) {

	initEthereumConf(tokenAddr, payChanAddr, apiUrl)
	_appInstance.sysImp = si
	_appInstance.dataImp = di

	if err := utils.TouchDir(baseDir); err != nil {
		errStr := fmt.Sprintf("touch dir(%s) err:%s", baseDir, err.Error())
		return ErrCreateDir, C.CString(errStr)
	}

	walletPath := filepath.Join(baseDir, string(filepath.Separator), WalletFile)
	receiptPath := filepath.Join(baseDir, string(filepath.Separator), ReceiptDataBase)
	cachePath := filepath.Join(baseDir, string(filepath.Separator), BlockDataBase)

	_appInstance.conf.baseDir = baseDir
	_appInstance.conf.receiptPath = receiptPath
	_appInstance.conf.walletPath = walletPath
	_appInstance.conf.cachePath = cachePath
	fmt.Println(_appInstance.conf.String())

	pp, err := payment.InitProtocol(walletPath, receiptPath, _appInstance)
	if err != nil {
		return ErrInitProtocol, C.CString(err.Error())
	}
	_appInstance.protocol = pp

	cc, err := payment.InitBlockDataCache(cachePath, _appInstance)
	if err != nil {
		return ErrInitDataCache, C.CString(err.Error())
	}
	_appInstance.dataSrv = cc
	return Success, nil
}

//export syncAppDataFromBlockChain
func syncAppDataFromBlockChain() {
	_appInstance.protocol.SyncWalletBalance()
	_appInstance.dataSrv.SyncPacketMarket()
	ab := _appInstance.protocol.AccBookInfo()
	if ab.MainAddr != "" {
		_appInstance.dataSrv.SyncMyChannelDetails(ab.MainAddr)
	}
}

//export startService
func startService(srvAddr, auth, minerPoolAddr string) (int, *C.char) {
	srv, err := proxy.NewProxyService(srvAddr, nil)
	if err != nil {
		return ErrInitVpnService, C.CString(err.Error())
	}
	_appInstance.service = srv

	if !_appInstance.protocol.IsPayChannelOpen(minerPoolAddr) {

		pool, ok := _appInstance.dataSrv.PoolDetails[minerPoolAddr]
		if !ok {
			return ErrNoSuchPool, C.CString("No such miner pool")
		}

		if err := _appInstance.protocol.OpenPayChannel(_appInstance.err, pool, auth); err != nil {
			return ErrOpenPayChannel, C.CString(err.Error())
		}
	}

	go srv.Accepting(_appInstance.err, proxy.Socks5Target, _appInstance.protocol)
	ret := <-_appInstance.err

	return ErrVpnServiceExit, C.CString(ret.Error())
}

//export stopService
func stopService() {
	_appInstance.err <- fmt.Errorf("stopped by user")
}
