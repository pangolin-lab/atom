package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

func ApproveToSpend(tokenNo *big.Int, key *ecdsa.PrivateKey) (string, error) {
	_, token, err := tokenConn()
	if err != nil {
		fmt.Println("[BuyPacket]: tokenConn err:", err.Error())
		return "", err
	}
	transactOpts := bind.NewKeyedTransactor(key)
	tx, err := token.Approve(transactOpts, common.HexToAddress(Conf.MicroPaySys), tokenNo)
	if err != nil {
		fmt.Println("[BuyPacket]: Approve err:", err.Error())
		return "", err
	}
	return tx.Hash().String(), nil
}

func BuyPacket(userAddr, poolAddr string, tokenNo *big.Int, key *ecdsa.PrivateKey) (string, error) {
	uAddr := common.HexToAddress(userAddr)
	pAddr := common.HexToAddress(poolAddr)

	conn, err := connect()
	if err != nil {
		fmt.Println("[BuyPacket]: BuyPacket err:", err.Error())
		return "", err
	}

	transactOpts := bind.NewKeyedTransactor(key)
	tx, err := conn.BuyPacket(transactOpts, uAddr, tokenNo, pAddr)
	if err != nil {
		fmt.Println("[BuyPacket]: BuyPacket err:", err.Error())
		return "", err
	}
	return tx.Hash().Hex(), nil
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
