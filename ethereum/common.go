package ethereum

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"strings"
)

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

func CreateEthAccount2(password, directory string) string {
	ks := keystore.NewKeyStore(directory, keystore.StandardScryptN, keystore.StandardScryptP)
	acc, err := ks.NewAccount(password)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	fmt.Println(acc.Address.Hex())
	fmt.Println(acc.URL.Path)

	path := acc.URL.Path
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
