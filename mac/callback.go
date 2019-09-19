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

func (app *MacApp) DataSyncedSuccess(typ int) {
	C.bridge_data_func(app.dataImp, C.int(typ), nil)
}
func (app *MacApp) DataSyncedFailed(err error) {
	C.bridge_data_func(app.dataImp, -1, C.CString(err.Error()))
}
