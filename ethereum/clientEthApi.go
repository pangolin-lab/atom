package ethereum

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pangolink/go-node/account"
	com "github.com/pangolink/miner-pool/common"
	"github.com/pangolink/miner-pool/eth/generated"
	"math"
	"math/big"
	"time"
)

var Conf = com.TestNet

func ConvertByDecimal(val *big.Int) float64 {
	fVal := new(big.Float)
	fVal.SetString(val.String())
	ethValue := new(big.Float).Quo(fVal, big.NewFloat(math.Pow10(com.TokenDecimal)))
	ret, _ := ethValue.Float64()
	return ret
}

func ConvertByFloat(val float64) *big.Int {
	valF := big.NewFloat(val)
	dec := big.NewFloat(math.Pow10(com.TokenDecimal))

	valF = valF.Mul(valF, dec)
	tn := new(big.Int)
	valF.Int(tn)
	return tn
}

func connect() (*generated.MicroPaySystem, error) {
	conn, err := ethclient.Dial(Conf.EthApiUrl)
	if err != nil {
		return nil, err
	}

	return generated.NewMicroPaySystem(common.HexToAddress(Conf.MicroPaySys), conn)
}

func tokenConn() (*ethclient.Client, *generated.PPNToken, error) {
	conn, err := ethclient.Dial(Conf.EthApiUrl)
	if err != nil {
		return nil, nil, err
	}
	token, err := generated.NewPPNToken(common.HexToAddress(Conf.Token), conn)
	return conn, token, err
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

func MySubPools(addr string) string {

	conn, err := connect()
	if err != nil {
		fmt.Println("[MySubPools]: connect err:", err.Error())
		return ""
	}

	arr, err := conn.AllMySubPools(nil, common.HexToAddress(addr))
	if err != nil {
		fmt.Println("[MySubPools]: AllMySubPools err:", err.Error())
		return ""
	}

	bytes, err := json.Marshal(arr)
	if err != nil {
		fmt.Println("[MySubPools]: marshal pool addresses arrays err:", err.Error())
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
		fmt.Println("[MySubPoolsWithDetails]: AllMySubPools err:", err.Error())
		return ""
	}

	poolArr := make([]PayChannel, 0)
	for i := 0; i < len(arr); i++ {
		poolAddr := arr[i]
		detail, err := conn.MicroPaymentChannels(nil, myAddr, poolAddr)
		if err != nil {
			fmt.Println("[MySubPoolsWithDetails]: MicroPaymentChannels err:", err.Error())
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
		fmt.Println("[MySubPoolsWithDetails]: marshal pool with details arrays err:", err.Error())
		return ""
	}

	return string(b)
}

func BuyPacket(userAddr, poolAddr string, tokenNo float64, key *ecdsa.PrivateKey) (string, error) {
	client, conn, err := tokenConn()
	if err != nil {
		fmt.Println("[BuyPacket]: tokenConn err:", err.Error())
		return "", err
	}
	uAddr := common.HexToAddress(userAddr)
	pAddr := common.HexToAddress(poolAddr)
	tn := ConvertByFloat(tokenNo)
	a := QueryApproved(uAddr)

	transactOpts := bind.NewKeyedTransactor(key)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Println("[BuyPacket]: SuggestGasPrice err:", err.Error())
		return "", err
	}
	transactOpts.GasPrice = gasPrice.Mul(gasPrice, big.NewInt(2))
	if a == nil || a.Cmp(tn) < 0 {
		tx, err := conn.Approve(transactOpts, common.HexToAddress(Conf.MicroPaySys), tn)
		if err != nil {
			fmt.Println("[BuyPacket]: Approve err:", err.Error())
			return "", err
		}

		fmt.Println("[BuyPacket]: Approve success:", tx.Hash().Hex())

		for {
			receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				<-time.After(time.Second)
				fmt.Println("[BuyPacket]: TransactionReceipt unfinished:", err.Error())
				continue
			}
			if receipt.Status != 1 {
				return "", fmt.Errorf("approve token balance failed")
			} else {
				fmt.Println("[BuyPacket]: approve success:", receipt.BlockHash.Hex())
				break
			}
		}
	}

	mConn, err := connect()
	if err != nil {
		fmt.Println("[BuyPacket]: connect err:", err.Error())
		return "", err
	}

	tx, err := mConn.BuyPacket(transactOpts, uAddr, tn, pAddr)
	if err != nil {
		fmt.Println("[BuyPacket]: BuyPacket err:", err.Error())
		return "", err
	}
	return tx.Hash().Hex(), nil
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

func TransferEth(target string, tokenNo float64, privateKey *ecdsa.PrivateKey) (string, error) {

	conn, err := ethclient.Dial(Conf.EthApiUrl)
	if err != nil {
		fmt.Println("[TransferEth]: Dial err:", err.Error())
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("[TransferEth]: cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return "", err
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := conn.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		fmt.Println("[TransferEth]: PendingNonceAt err:", err.Error())
		return "", err
	}

	value := ConvertByFloat(tokenNo) // in wei (1 eth)
	gasLimit := uint64(21000)        // in units
	gasPrice, err := conn.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Println("[TransferEth]: SuggestGasPrice err:", err.Error())
		return "", err
	}

	var data []byte
	tx := types.NewTransaction(nonce, common.HexToAddress(target), value, gasLimit, gasPrice, data)

	chainID, err := conn.NetworkID(context.Background())
	if err != nil {
		fmt.Println("[TransferEth]: NetworkID err:", err.Error())
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		fmt.Println("[TransferEth]: SignTx err:", err.Error())
		return "", err
	}

	err = conn.SendTransaction(context.Background(), signedTx)
	if err != nil {
		fmt.Println("[TransferEth]: SendTransaction err:", err.Error())
		return "", err
	}
	return signedTx.Hash().Hex(), nil
}

func TransferLinToken(target string, tokenNo float64, key *ecdsa.PrivateKey) (string, error) {

	_, conn, err := tokenConn()
	if err != nil {
		fmt.Println("[TransferLinToken]: tokenConn err:", err.Error())
		return "", err
	}
	opts := bind.NewKeyedTransactor(key)
	val := ConvertByFloat(tokenNo)

	fmt.Printf("\n----->%.2f", ConvertByDecimal(val))

	tx, err := conn.Transfer(opts, common.HexToAddress(target), val)
	if err != nil {
		fmt.Println("[TransferLinToken]: Transfer err:", err.Error())
		return "", err
	}

	return tx.Hash().Hex(), nil
}
