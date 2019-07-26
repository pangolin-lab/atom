package eth

import "C"
import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/proton-lab/proton/account"
	"github.com/proton-lab/proton/service"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"strings"
)

func freeManager() (*ethclient.Client, *service.ProtonManager, error) {
	conn, err := ethclient.Dial(service.EthereNetworkAPI)
	if err != nil {
		fmt.Printf("\nDial up infura failed:%s", err)
		return nil, nil, err
	}

	manager, err := service.NewProtonManager(common.HexToAddress(service.ProtonManagerContractAddress), conn)
	if err != nil {
		fmt.Printf("\nCreate Proton Manager err:%s", err)
		conn.Close()
		return nil, nil, err
	}

	return conn, manager, nil
}

func payableManager(cipherKey, password string) (*ethclient.Client, *service.ProtonManager, *bind.TransactOpts, error) {

	conn, err := ethclient.Dial(service.EthereNetworkAPI)
	if err != nil {
		fmt.Printf("\nDial up infura failed:%s", err)
		return nil, nil, nil, err
	}

	manager, err := service.NewProtonManager(common.HexToAddress(service.ProtonManagerContractAddress), conn)
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

func ConvertByDecimal(val *big.Int) float64 {
	fVal := new(big.Float)
	fVal.SetString(val.String())
	ethValue := new(big.Float).Quo(fVal, big.NewFloat(math.Pow10(18)))
	ret, _ := ethValue.Float64()
	return ret
}

func CreateEthAccount(password, directory string) string {

	ks := keystore.NewKeyStore(directory, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.NewAccount(password)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	fmt.Println(account.Address.Hex())
	fmt.Println(account.URL.Path)
	return account.Address.Hex()
}

func CreateEthAccount2(password, directory string) string {

	ks := keystore.NewKeyStore(directory, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.NewAccount(password)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	fmt.Println(account.Address.Hex())
	fmt.Println(account.URL.Path)

	path := account.URL.Path
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	buffer := make([]byte, 10240)
	n, err := file.Read(buffer)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	file.Close()
	os.Remove(path)

	return string(buffer[:n])
}

func ImportEthAccount(file, dir, password string) string {

	ks := keystore.NewKeyStore(dir, keystore.StandardScryptN, keystore.StandardScryptP)
	jsonBytes, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	account, err := ks.Import(jsonBytes, password, password)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	fmt.Println(account.Address.Hex())
	return account.Address.Hex()
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
