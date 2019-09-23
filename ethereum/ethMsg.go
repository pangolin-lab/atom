package ethereum

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

func TokenBalance(address string) (*big.Int, *big.Int, *big.Int) {
	conn, err := connect()
	if err != nil {
		fmt.Print(err)
		return nil, nil, nil
	}

	tokenB, ethB, approved, err := conn.TokenBalance(nil, common.HexToAddress(address))
	if err != nil {
		fmt.Print(err)
		return nil, nil, nil
	}
	return ethB, tokenB, approved
	//return ConvertByDecimal(tokenB), ConvertByDecimal(ethB)
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

func PoolDetails(addr common.Address) (*PoolDetail, error) {

	conn, err := connect()
	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	d, err := conn.MinerPools(nil, addr)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	details := &PoolDetail{
		MainAddr:     d.MainAddr.String(),
		Payer:        d.Payer.String(),
		GuaranteedNo: ConvertByDecimal(d.GuaranteedNo),
		ShortName:    d.ShortName,
		DetailInfos:  d.DetailInfos,
		Seeds:        d.Seeds,
	}
	return details, nil
}
func PoolListWithDetails() string {
	conn, err := connect()
	if err != nil {
		fmt.Println("[PoolListWithDetails]: connect err:", err.Error())
		return ""
	}

	addrList, err := conn.GetPoolAddress(nil)
	if err != nil {
		fmt.Println("[PoolListWithDetails]: GetPoolAddress err:", err)
		return ""
	}
	arr := make([]PoolDetail, 0)
	for i := 0; i < len(addrList); i++ {
		d, err := conn.MinerPools(nil, addrList[i])
		if err != nil {
			fmt.Println("[PoolListWithDetails]: MinerPools err:", err)
			continue
		}

		details := PoolDetail{
			MainAddr:     d.MainAddr.String(),
			Payer:        d.Payer.String(),
			GuaranteedNo: ConvertByDecimal(d.GuaranteedNo),
			ShortName:    d.ShortName,
			DetailInfos:  d.DetailInfos,
			Seeds:        d.Seeds,
		}

		arr = append(arr, details)
	}
	if len(arr) == 0 {
		fmt.Println("[PoolListWithDetails]: no valid pool items")
		return ""
	}
	buf, err := json.Marshal(arr)
	if err != nil {
		fmt.Println("[PoolListWithDetails]: Marshal miner pool detail array err:", err)
		return ""
	}
	return string(buf)
}

func MySubPools(addr string) ([]common.Address, error) {
	conn, err := connect()
	if err != nil {
		fmt.Println("[MySubPools]: connect err:", err.Error())
		return nil, err
	}
	return conn.AllMySubPools(nil, common.HexToAddress(addr))
}

func MyChannelWithDetails(addr string) ([]PayChannel, error) {
	conn, err := connect()
	if err != nil {
		fmt.Println("[Atom]: connect err:", err.Error())
		return nil, err
	}

	myAddr := common.HexToAddress(addr)
	arr, err := conn.AllMySubPools(nil, myAddr)
	if err != nil {
		fmt.Println("[MyChannelWithDetails]: AllMySubPools err:", err.Error())
		return nil, err
	}

	poolArr := make([]PayChannel, 0)
	for i := 0; i < len(arr); i++ {
		poolAddr := arr[i]
		detail, err := conn.MicroPaymentChannels(nil, myAddr, poolAddr)
		if err != nil {
			fmt.Println("[MyChannelWithDetails]: MicroPaymentChannels err:", err.Error())
			continue
		}

		d := PayChannel{
			MainAddr:      poolAddr.String(),
			RemindTokens:  ConvertByDecimal(detail.RemindTokens),
			RemindPackets: detail.RemindPackets.Int64(),
			Expiration:    detail.Expiration.Int64(),
		}
		poolArr = append(poolArr, d)
	}
	return poolArr, nil
}

func QueryApproved(address common.Address) *big.Int {

	_, conn, err := tokenConn()
	if err != nil {
		fmt.Println("[QueryApproved]: tokenConn err:", err.Error())
		return nil
	}

	a, e := conn.Allowance(nil, address, common.HexToAddress(Conf.MicroPaySys))
	if e != nil {
		return nil
	}
	fmt.Println(a.String())
	return a
}

func QueryMicroPayPrice() *big.Int {
	conn, err := connect()
	if err != nil {
		fmt.Println("[QueryMicroPayPrice]: connect err:", err.Error())
		return nil
	}

	p, err := conn.PacketPrice(nil)
	if err != nil {
		fmt.Println("[QueryMicroPayPrice]: PacketPrice err:", err.Error())
		return nil
	}
	return p
}

func MarketDataVersion() uint32 {
	conn, err := connect()
	if err != nil {
		fmt.Println("[TransferLinToken]: connect err:", err.Error())
		return 0
	}
	ver, err := conn.MinerPoolVersion(nil)
	if err != nil {
		fmt.Println("[TransferLinToken]: MinerPoolVersion err:", err.Error())
		return 0
	}

	return ver
}

func TxStatus(tx common.Hash) bool {

	client, err := ethclient.Dial(Conf.EthApiUrl)
	if err != nil {
		fmt.Println("[BuyPacket]: connect err:", err.Error())
		return false
	}

	receipt, err := client.TransactionReceipt(context.Background(), tx)
	if err != nil {
		fmt.Println("[BuyPacket]: query receipt err:", err.Error())
		return false
	}
	return receipt.Status == 1
}
