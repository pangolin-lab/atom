package ethereum

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	com "github.com/pangolink/miner-pool/common"
	"github.com/pangolink/miner-pool/eth/generated"
	"math"
	"math/big"
)

var Conf *com.EthereumConfig = com.TestNet

func ConvertByDecimal(val *big.Int) float64 {
	fVal := new(big.Float)
	fVal.SetString(val.String())
	ethValue := new(big.Float).Quo(fVal, big.NewFloat(math.Pow10(18)))
	ret, _ := ethValue.Float64()
	return ret
}

func connect() (*generated.MicroPaySystem, error) {
	conn, err := ethclient.Dial(Conf.EthApiUrl)
	if err != nil {
		return nil, err
	}

	return generated.NewMicroPaySystem(common.HexToAddress(Conf.MicroPaySys), conn)
}

func FlowDataBalance(userAddr, poolAddr string) int64 {
	return 0
}

func BuyFlowData(key *ecdsa.PrivateKey) (string, error) {

	//conn, err := connect()
	//if err != nil{
	//	return "", err
	//}
	//transactOpts := bind.NewKeyedTransactor(key)
	//
	//tx, err := conn.BuyPacket(transactOpts, [32]byte, )

	return "", nil
}

func TokenBalance(address string) (float64, float64) {
	conn, err := connect()
	if err != nil {
		fmt.Print(err)
		return 0, 0
	}

	tokenB, ethB, err := conn.TokenBalance(nil, common.HexToAddress(address))
	if err != nil {
		fmt.Print(err)
		return 0, 0
	}

	return ConvertByDecimal(tokenB), ConvertByDecimal(ethB)
}

func PoolAddressList() []common.Address {
	conn, err := connect()
	if err != nil {
		fmt.Print(err)
		return nil
	}

	addArr, err := conn.GetPoolAddrees(nil)
	if err != nil {
		fmt.Print(err)
		return nil
	}

	return addArr
}

func PoolDetails(addr string) string {

	conn, err := connect()
	if err != nil {
		fmt.Print(err)
		return ""
	}

	pool, err := conn.MinerPools(nil, common.HexToAddress(addr))
	if err != nil {
		fmt.Print(err)
		return ""
	}

	buf, err := json.Marshal(pool)
	if err != nil {
		fmt.Print(err)
		return ""
	}

	return string(buf)
}

func PoolListWithDetails() string {

	conn, err := connect()
	if err != nil {
		fmt.Print(err)
		return ""
	}

	addrList, err := conn.GetPoolAddrees(nil)
	if err != nil {
		fmt.Print(err)
		return ""
	}

	arr := make([]interface{}, len(addrList))
	for i := 0; i < len(addrList); i++ {
		details, err := conn.MinerPools(nil, addrList[i])
		if err != nil {
			fmt.Print(err)
			continue
		}
		arr = append(arr, details)
	}
	buf, err := json.Marshal(arr)
	if err != nil {
		fmt.Print(err)
		return ""
	}
	return string(buf)
}
