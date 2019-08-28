package ethereum

import (
	"github.com/pangolink/miner-pool/common"
	"github.com/pangolink/miner-pool/eth"
	"math"
	"math/big"
)

func ConvertByDecimal(val *big.Int) float64 {
	fVal := new(big.Float)
	fVal.SetString(val.String())
	ethValue := new(big.Float).Quo(fVal, big.NewFloat(math.Pow10(18)))
	ret, _ := ethValue.Float64()
	return ret
}

var EthConf *common.EthereumConfig = nil

func TokenBalance(address string) (int64, int64) {
	client := eth.NewClientApi(EthConf)
	return client.TokenBalance(address)
}
