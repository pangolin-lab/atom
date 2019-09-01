package ethereum

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pangolink/go-node/account"
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

type PoolDetail struct {
	MainAddr     string
	Payer        string
	SubAddr      string
	GuaranteedNo float64
	ID           int
	PoolType     uint8
	ShortName    string
	DetailInfos  string
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

	arr := make([]PoolDetail, 0)
	for i := 0; i < len(addrList); i++ {
		d, err := conn.MinerPools(nil, addrList[i])
		if err != nil {
			fmt.Print(err)
			continue
		}

		details := PoolDetail{
			MainAddr:     d.MainAddr.Hex(),
			Payer:        d.Payer.Hex(),
			SubAddr:      account.ConvertToID2(d.SubAddr).String(),
			GuaranteedNo: ConvertByDecimal(d.GuaranteedNo),
			ID:           int(d.ID.Int64()),
			PoolType:     d.PoolType,
			ShortName:    d.ShortName,
			DetailInfos:  d.DetailInfos,
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
