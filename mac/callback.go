package main

/*
#include "callback.h"

void bridge_data_func(BlockChainDataSyncNotifier f , int t, char* v){
	f(t, v);
}

void bridge_sys_func(SystemActionCallBack f, int t, char* v){
	f(t, v);
}
*/
import "C"
import "github.com/pangolin-lab/atom/payment"

func (app *MacApp) SubPoolDataSynced() {
	C.bridge_data_func(app.dataImp, C.SubPoolSynced, nil)
}

func (app *MacApp) MarketPoolDataSynced() {
	C.bridge_data_func(app.dataImp, C.MarketPoolSynced, nil)
}

func (app *MacApp) DataSyncedFailed(err error) {
	C.bridge_data_func(app.dataImp, -1, C.CString(err.Error()))
}

func (app *MacApp) WalletBalanceSynced() {
	C.bridge_sys_func(app.sysImp, C.BalanceSynced, nil)
}

func (app *MacApp) NotifyApproveToSystem(tx string) {
	C.bridge_sys_func(app.sysImp, C.ApproveToSpendCoin, C.CString(tx))
}

func (app *MacApp) ReInitApp() error {

	conf := app.conf
	app.protocol.Finalized()
	app.dataSrv.Finalized()

	p, e := payment.InitProtocol(conf.walletPath, conf.receiptPath, app)
	if e != nil {
		return e
	}
	app.protocol = p

	ab := app.protocol.SyncWalletData()
	d, e := payment.InitBlockDataCache(conf.cachePath, ab.MainAddr, app)
	if e != nil {
		p.Finalized()
		return e
	}
	app.dataSrv = d

	return nil
}
