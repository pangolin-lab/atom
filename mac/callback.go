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
