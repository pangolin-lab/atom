package ethereum

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	com "github.com/pangolink/miner-pool/common"
	"github.com/pangolink/miner-pool/eth/generated"
	"math"
	"math/big"
)

var Conf = com.TestNet

type PoolDetail struct {
	MainAddr     string
	Payer        string
	GuaranteedNo *big.Int
	ShortName    string
	DetailInfos  string
	Seeds        string
}

type ChannelDetail struct {
	MainAddr      string
	RemindPackets *big.Int
	Expiration    *big.Int
}

func ConvertByDecimal(val *big.Int) float64 {
	if val == nil {
		return 0
	}

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
