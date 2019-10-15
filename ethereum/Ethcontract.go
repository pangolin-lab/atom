package ethereum

import "C"
import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pangolink/pangolin-node/account"
	"github.com/pangolink/pangolin-node/service/ethInterface"
	"io/ioutil"
	"math"
	"math/big"
	"strings"
)

func freeManager() (*ethclient.Client, *ethInterface.HopManager, error) {
	conn, err := ethclient.Dial(ethInterface.EthereNetworkAPI)
	if err != nil {
		fmt.Printf("\nDial up infura failed:%s", err)
		return nil, nil, err
	}

	manager, err := ethInterface.NewHopManager(common.HexToAddress(ethInterface.ManagerContractAddress), conn)
	if err != nil {
		fmt.Printf("\nCreate Proton Manager err:%s", err)
		conn.Close()
		return nil, nil, err
	}

	return conn, manager, nil
}

func payableManager(cipherKey, password string) (*ethclient.Client, *ethInterface.HopManager, *bind.TransactOpts, error) {

	conn, err := ethclient.Dial(ethInterface.EthereNetworkAPI)
	if err != nil {
		fmt.Printf("\nDial up infura failed:%s", err)
		return nil, nil, nil, err
	}

	manager, err := ethInterface.NewHopManager(common.HexToAddress(ethInterface.ManagerContractAddress), conn)
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
	ethAddr, _, _, err := manager.Check(nil, arr)
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

	ethBalance, protonBalance, protonNo, err := manager.BindingInfo(nil, common.HexToAddress(ethAddr))
	if err != nil {
		fmt.Printf("\n CheckBinder err:%s", err)
		return nil, nil, 0
	}

	fmt.Printf("\n ETH=%d Proton=%d NO=%d\n", ethBalance, protonBalance, protonNo)
	return ethBalance, protonBalance, int(protonNo.Int64())
}

func ConvertByDecimal(val *big.Int) float64 {
	fVal := new(big.Float)
	fVal.SetString(val.String())
	ethValue := new(big.Float).Quo(fVal, big.NewFloat(math.Pow10(18)))
	ret, _ := ethValue.Float64()
	return ret
}

func CreateEthAccount(password, directory string) string {

	ks := keystore.NewKeyStore(directory, keystore.StandardScryptN, keystore.StandardScryptP)
	acc, err := ks.NewAccount(password)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	fmt.Println(acc.Address.Hex())
	fmt.Println(acc.URL.Path)
	return acc.Address.Hex()
}

func ImportEthAccount(file, dir, password string) string {

	ks := keystore.NewKeyStore(dir, keystore.StandardScryptN, keystore.StandardScryptP)
	jsonBytes, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	acc, err := ks.Import(jsonBytes, password, password)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	fmt.Println(acc.Address.Hex())
	return acc.Address.Hex()
}

func VerifyEthAccount(cipherTxt, passphrase string) bool {

	keyin := strings.NewReader(cipherTxt)
	json, err := ioutil.ReadAll(keyin)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if _, err := keystore.DecryptKey(json, passphrase); err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func BindProtonAddr(protonAddr, cipherKey, password string) (string, error) {

	conn, manager, auth, err := payableManager(cipherKey, password)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	arr := account.ID(protonAddr).ToArray()
	tx, err := manager.Bind(auth, arr)
	if err != nil {
		return "", err
	}

	fmt.Printf("\nTransfer pending: 0x%x for proton addr:%s \n", tx.Hash(), hex.EncodeToString(arr[:]))
	return tx.Hash().String(), err
}

func UnbindProtonAddr(protonAddr, cipherKey, password string) (string, error) {

	var conn, manager, auth, err = payableManager(cipherKey, password)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	arr := account.ID(protonAddr).ToArray()
	tx, err := manager.Unbind(auth, arr)
	if err != nil {
		return "", err
	}

	fmt.Printf("\nTransfer pending: 0x%x for Proton addr:%s \n", tx.Hash(), hex.EncodeToString(arr[:]))
	return tx.Hash().String(), err
}
