package main

import "C"

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolink/miner-pool/account"
)

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

//export MySubPools
func MySubPools(addr string) *C.char {
	jsonStr := ethereum.MySubPools(addr)
	return C.CString(jsonStr)
}

//export MySubPoolsWithDetails
func MySubPoolsWithDetails(addr string) *C.char {
	jsonStr := ethereum.MySubPoolsWithDetails(addr)
	return C.CString(jsonStr)
}

//export BuyPacket
func BuyPacket(userAddr, poolAddr, passPhrase, cipherTxt string, tokenNo float64) (*C.char, *C.char) {

	w, e := account.DecryptWallet([]byte(cipherTxt), passPhrase)
	if e != nil {
		return C.CString(""), C.CString(e.Error())
	}

	tx := ethereum.BuyPacket(userAddr, poolAddr, tokenNo, w.SignKey())
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
