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

var Conf = com.TestNet

func ConvertByDecimal(val *big.Int) float64 {
	fVal := new(big.Float)
	fVal.SetString(val.String())
	ethValue := new(big.Float).Quo(fVal, big.NewFloat(math.Pow10(com.TokenDecimal)))
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

	addArr, err := conn.GetPoolAddress(nil)
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
	MainAddr     common.Address
	Payer        common.Address
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
		fmt.Println("[Atom]: connect err:", err.Error())
		return ""
	}

	addrList, err := conn.GetPoolAddress(nil)
	if err != nil {
		fmt.Println("[Atom]: GetPoolAddress err:", err)
		return ""
	}
	arr := make([]PoolDetail, 0)
	for i := 0; i < len(addrList); i++ {
		d, err := conn.MinerPools(nil, addrList[i])
		if err != nil {
			fmt.Println("[Atom]: MinerPools err:", err)
			continue
		}

		details := PoolDetail{
			MainAddr:     d.MainAddr,
			Payer:        d.Payer,
			SubAddr:      account.ConvertToID2(d.SubAddr).String(),
			GuaranteedNo: ConvertByDecimal(d.GuaranteedNo),
			ID:           int(d.ID),
			PoolType:     d.PoolType,
			ShortName:    d.ShortName,
			DetailInfos:  d.DetailInfos,
		}

		arr = append(arr, details)
	}
	if len(arr) == 0 {
		fmt.Println("[Atom]: no valid pool items")
		return ""
	}
	buf, err := json.Marshal(arr)
	if err != nil {
		fmt.Println("[Atom]: Marshal miner pool detail array err:", err)
		return ""
	}
	return string(buf)
}

func MySubPools(addr string) string {

	conn, err := connect()
	if err != nil {
		fmt.Println("[Atom]: connect err:", err.Error())
		return ""
	}

	arr, err := conn.AllMySubPools(nil, common.HexToAddress(addr))
	if err != nil {
		fmt.Println("[Atom]: AllMySubPools err:", err.Error())
		return ""
	}

	bytes, err := json.Marshal(arr)
	if err != nil {
		fmt.Println("[Atom]: marshal pool addresses arrays err:", err.Error())
		return ""
	}

	return string(bytes)
}

type PayChannel struct {
	MainAddr      common.Address
	PayerAddr     common.Address
	RemindTokens  float64
	RemindPackets int64
	Expiration    int64
}

func MySubPoolsWithDetails(addr string) string {
	conn, err := connect()
	if err != nil {
		fmt.Println("[Atom]: connect err:", err.Error())
		return ""
	}
	myAddr := common.HexToAddress(addr)
	arr, err := conn.AllMySubPools(nil, myAddr)
	if err != nil {
		fmt.Println("[Atom]: AllMySubPools err:", err.Error())
		return ""
	}

	poolArr := make([]PayChannel, 0)
	for i := 0; i < len(arr); i++ {
		poolAddr := arr[i]
		detail, err := conn.MicroPaymentChannels(nil, myAddr, poolAddr)
		if err != nil {
			fmt.Println("[Atom]: MicroPaymentChannels err:", err.Error())
			continue
		}

		d := PayChannel{
			MainAddr:      detail.MainAddr,
			PayerAddr:     detail.PayerAddr,
			RemindTokens:  ConvertByDecimal(detail.RemindTokens),
			RemindPackets: detail.RemindPackets.Int64(),
			Expiration:    detail.Expiration.Int64(),
		}
		poolArr = append(poolArr, d)
	}

	b, err := json.Marshal(poolArr)
	if err != nil {
		fmt.Println("[Atom]: marshal pool with details arrays err:", err.Error())
		return ""
	}

	return string(b)
}
