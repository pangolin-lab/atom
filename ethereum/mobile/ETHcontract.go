package mobile

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"os"
)
import "github.com/ethereum/go-ethereum/mobile"

type EthContractI struct {
	*geth.EthereumClient
}

const EthereNetworkAPI = "https://ropsten.infura.io/v3/8b8db3cca50a4fcf97173b7619b1c4c3"

func newCli() *EthContractI {
	c, err := geth.NewEthereumClient(EthereNetworkAPI)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	ec := &EthContractI{
		c,
	}
	return ec
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
