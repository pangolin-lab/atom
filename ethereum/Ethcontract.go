package ethereum

import "C"
import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/proton-lab/proton-node/account"
	"github.com/proton-lab/proton-node/service/ethInterface"
	"github.com/proton-lab/proton-node/service/rpcMsg"
	"math/big"
	"strings"
)

func freeManager() (*ethclient.Client, *ethInterface.ProtonManager, error) {
	conn, err := ethclient.Dial(rpcMsg.EthereNetworkAPI)
	if err != nil {
		fmt.Printf("\nDial up infura failed:%s", err)
		return nil, nil, err
	}

	manager, err := ethInterface.NewProtonManager(common.HexToAddress(ethInterface.ProtonManagerContractAddress), conn)
	if err != nil {
		fmt.Printf("\nCreate Proton Manager err:%s", err)
		conn.Close()
		return nil, nil, err
	}

	return conn, manager, nil
}

func payableManager(cipherKey, password string) (*ethclient.Client, *ethInterface.ProtonManager, *bind.TransactOpts, error) {

	conn, err := ethclient.Dial(rpcMsg.EthereNetworkAPI)
	if err != nil {
		fmt.Printf("\nDial up infura failed:%s", err)
		return nil, nil, nil, err
	}

	manager, err := ethInterface.NewProtonManager(common.HexToAddress(ethInterface.ProtonManagerContractAddress), conn)
	if err != nil {
		fmt.Printf("\nCreate Proton Manager err:%s", err)
		conn.Close()
		return nil, nil, nil, err
	}

	auth, err := bind.NewTransactor(strings.NewReader(cipherKey), password)
	if err != nil {
		conn.Close()
		return nil, nil, nil, err
	}
	return conn, manager, auth, nil
}

func CheckProtonAddr(protonAddr string) string {

	conn, manager, err := freeManager()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer conn.Close()

	arr := account.ID(protonAddr).ToArray()
	fmt.Printf("\nQuery proton [%s] ehtereum address (%s)", protonAddr, hex.EncodeToString(arr[:]))
	ethAddr, _, _, err := manager.CheckProtonAddress(nil, arr)
	if err != nil {
		fmt.Printf("\n CheckProtonAddress err:%s", err)
		return ""
	}
	return ethAddr.Hex()
}

func BalanceOfEthAddr(ethAddr string) (*big.Int, *big.Int, int) {

	conn, manager, err := freeManager()
	if err != nil {
		fmt.Println(err)
		return nil, nil, 0
	}
	defer conn.Close()

	ethBalance, protonBalance, protonNo, err := manager.CheckBinder(nil, common.HexToAddress(ethAddr))
	if err != nil {
		fmt.Printf("\n CheckBinder err:%s", err)
		return nil, nil, 0
	}

	fmt.Printf("ETH=%d Proton=%d NO=%d", ethBalance, protonBalance, protonNo)
	return ethBalance, protonBalance, int(protonNo.Int64())
}

func BindProtonAddr(protonAddr, cipherKey, password string) (string, error) {

	conn, manager, auth, err := payableManager(cipherKey, password)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	arr := account.ID(protonAddr).ToArray()
	tx, err := manager.BindProtonAddress(auth, arr)
	if err != nil {
		return "", err
	}

	fmt.Printf("\nTransfer pending: 0x%x for proton addr:%s \n", tx.Hash(), hex.EncodeToString(arr[:]))
	return tx.Hash().String(), err
}

func UnbindProtonAddr(protonAddr, cipherKey, password string) (string, error) {

	conn, manager, auth, err := payableManager(cipherKey, password)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	arr := account.ID(protonAddr).ToArray()
	tx, err := manager.UnbindProtonAddress(auth, arr)
	if err != nil {
		return "", err
	}

	fmt.Printf("\nTransfer pending: 0x%x for Proton addr:%s \n", tx.Hash(), hex.EncodeToString(arr[:]))
	return tx.Hash().String(), err
}
