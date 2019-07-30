package ethereum

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/proton-lab/proton-node/account"
	"github.com/proton-lab/proton-node/service/ethInterface"
	"github.com/proton-lab/proton-node/service/rpcMsg"
	"math/big"
)

func freeSimpleManager() (*ethInterface.SimpleProtonManager, error) {
	conn, err := ethclient.Dial(rpcMsg.EthereNetworkAPI)
	if err != nil {
		fmt.Printf("\nDial up infura failed:%s", err)
		return nil, err
	}
	manager, err := ethInterface.NewSimpleProtonManager(common.HexToAddress(ethInterface.SimpleManagerContractAddress), conn)
	if err != nil {
		return nil, err
	}
	return manager, nil
}

func BasicBalance(ethAddr string) (*big.Int, int) {
	manager, err := freeSimpleManager()
	if err != nil {
		fmt.Println(err)
		return nil, 0
	}

	myAddress := common.HexToAddress(ethAddr)
	balance, no, err := manager.BasicBalance(nil, myAddress)
	if err != nil {
		fmt.Println(err)
		return nil, 0
	}

	return balance, int(no.Int64())
}

func BoundEth(protonAddr string) string {
	manager, err := freeSimpleManager()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	arr := account.ID(protonAddr).ToArray()
	address, err := manager.BoundedEthAddress(nil, arr)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return address.Hex()
}

func UnderMyAddress(ethAddr string, idx int) string {
	manager, err := freeSimpleManager()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	myAddress := common.HexToAddress(ethAddr)
	protonAddr, err := manager.AddressesUnderMyAccount(nil, myAddress, big.NewInt(int64(idx)))
	if err != nil {
		return ""
	}

	return account.ConvertToID2(protonAddr).String()
}

func UnderMyAddresses(ethAddr string, start, end int) []string {
	manager, err := freeSimpleManager()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	myAddress := common.HexToAddress(ethAddr)

	addrArr := make([]string, 0)
	for i := start; i < end; i++ {
		idx := big.NewInt(int64(i))
		protonAddr, err := manager.AddressesUnderMyAccount(nil, myAddress, idx)
		if err != nil {
			return addrArr
		}
		str := account.ConvertToID2(protonAddr).String()
		addrArr = append(addrArr, str)
	}
	return addrArr
}
