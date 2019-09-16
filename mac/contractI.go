package main

import "C"

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolink/miner-pool/account"
)

//export PoolDetails
func PoolDetails(addr string) *C.char {
	pd, err := ethereum.PoolDetails(common.HexToAddress(addr))
	if err != nil {
		fmt.Print(err)
		return C.CString("")
	}

	buf, err := json.Marshal(pd)
	if err != nil {
		fmt.Print(err)
		return C.CString("")
	}

	return C.CString(string(buf))
}

//export PoolListWithDetails
func PoolListWithDetails() *C.char {
	jsonStr := ethereum.PoolListWithDetails()
	return C.CString(jsonStr)
}

//export MyChannelWithDetails
func MyChannelWithDetails(addr string) *C.char {
	poolArr, err := ethereum.MyChannelWithDetails(addr)
	if err != nil {
		fmt.Println("[MyChannelWithDetails]: marshal pool with details arrays err:", err.Error())
		return C.CString("")
	}

	b, err := json.Marshal(poolArr)
	if err != nil {
		fmt.Println("[MyChannelWithDetails]: marshal pool with details arrays err:", err.Error())
		return C.CString("")
	}
	return C.CString(string(b))
}

//export BuyPacket
func BuyPacket(userAddr, poolAddr, passPhrase, cipher string, tokenNo float64) (*C.char, *C.char) {
	w, e := account.DecryptWallet([]byte(cipher), passPhrase)
	if e != nil {
		return C.CString(""), C.CString(e.Error())
	}

	tx, e := ethereum.BuyPacket(userAddr, poolAddr, tokenNo, w.SignKey())
	if e != nil {
		return C.CString(""), C.CString(e.Error())
	}
	fmt.Println(tx)
	return C.CString(tx), C.CString("")
}

//export QueryApproved
func QueryApproved(address string) float64 {

	no := ethereum.QueryApproved(common.HexToAddress(address))
	if no == nil {
		return 0.0
	}

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
