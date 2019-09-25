package main

import "C"

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pangolin-lab/atom/ethereum"
)

//export MyChannelWithDetails
func MyChannelWithDetails() *C.char {
	addrArr := _appInstance.dataSrv.MySubscribedPool
	if len(addrArr) == 0 {
		return nil
	}
	poolArr := _appInstance.dataSrv.LoadDetailsOfArr(addrArr)
	jsonStr, err := json.Marshal(poolArr)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return C.CString(string(jsonStr))
}

//export  SyncChannelWithDetails
func SyncChannelWithDetails(address string) {
	go _appInstance.dataSrv.SyncSubscribedPool(address)
}

//export AuthorizeTokenSpend
func AuthorizeTokenSpend(auth string, tokenNo float64) (*C.char, *C.char) {
	w, e := _appInstance.protocol.Wallet(auth)
	if e != nil {
		return nil, C.CString(e.Error())
	}
	tn := ethereum.ConvertByFloat(tokenNo)
	tx, err := ethereum.ApproveToSpend(tn, w.SignKey())
	if err != nil {
		return nil, C.CString(e.Error())
	}

	return C.CString(tx), nil
}

//export TxProcessStatus
func TxProcessStatus(tx string) bool {
	return ethereum.TxStatus(common.HexToHash(tx))
}

//export BuyPacket
func BuyPacket(userAddr, poolAddr, auth string, tokenNo float64) (*C.char, *C.char) {
	w, e := _appInstance.protocol.Wallet(auth)
	if e != nil {
		return nil, C.CString(e.Error())
	}

	tn := ethereum.ConvertByFloat(tokenNo)
	tx, e := ethereum.BuyPacket(userAddr, poolAddr, tn, w.SignKey())
	if e != nil {
		return nil, C.CString(e.Error())
	}
	fmt.Println(tx)
	return C.CString(tx), nil
}

//export QueryApproved
func QueryApproved(address string) float64 {
	no := ethereum.QueryApproved(common.HexToAddress(address))
	return ethereum.ConvertByDecimal(no)
}

//export QueryMicroPayPrice
func QueryMicroPayPrice() int64 {
	p := ethereum.QueryMicroPayPrice()
	if p == nil {
		return -1
	}
	return p.Int64()
}

//export PoolDetails
func PoolDetails(addr string) *C.char {
	pool, err := _appInstance.dataSrv.LoadPoolDetails(addr)
	if err != nil {
		fmt.Print(err)
		return nil
	}

	buf, err := json.Marshal(pool)
	if err != nil {
		fmt.Print(err)
		return nil
	}

	return C.CString(string(buf))
}

//export PoolInfosInMarket
func PoolInfosInMarket() *C.char {
	addrArr := _appInstance.dataSrv.PoolsInMarket
	if len(addrArr) == 0 {
		return nil
	}
	poolArr := _appInstance.dataSrv.LoadDetailsOfArr(addrArr)
	jsonStr, err := json.Marshal(poolArr)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return C.CString(string(jsonStr))
}

//export AsyncLoadMarketData
func AsyncLoadMarketData() {
	go _appInstance.dataSrv.SyncPacketMarket()
}
