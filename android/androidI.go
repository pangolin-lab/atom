package androidLib

import "C"
import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/proton-lab/autom/ethereum"
	"github.com/proton-lab/autom/ethereum/mobile"
	"github.com/proton-lab/autom/pipeProxy"
	"github.com/proton-lab/autom/tun2Pipe"
	"github.com/proton-lab/autom/wallet"
	"github.com/proton-lab/proton-node/account"
	"io/ioutil"
	"strings"
)

type VpnDelegate interface {
	tun2Pipe.VpnDelegate
	GetBootPath() string
}

const Separator = "@@@"

var _instance *pipeProxy.PipeProxy = nil
var proxyConf = &pipeProxy.ProxyConfig{}

func InitVPN(addr, cipher, url, boot, IPs string, d VpnDelegate) error {

	pt := func(fd uintptr) {
		d.ByPass(int32(fd))
	}

	proxyConf.WConfig = &wallet.WConfig{
		BCAddr:     addr,
		Cipher:     cipher,
		SettingUrl: url,
		Saver:      pt,
	}
	tun2Pipe.VpnInstance = d
	tun2Pipe.Protector = pt

	proxyConf.BootNodes = boot
	tun2Pipe.ByPassInst().Load(IPs)

	mis := proxyConf.FindBootServers(d.GetBootPath())
	if len(mis) == 0 {
		return fmt.Errorf("no valid boot strap node")
	}

	proxyConf.ServerId = mis[0]
	println(proxyConf.String())
	return nil
}

func SetupVpn(password, locAddr string) error {

	t2s, err := tun2Pipe.New(locAddr)
	if err != nil {
		return err
	}

	fmt.Println(proxyConf.String())

	w, err := wallet.NewWallet(proxyConf.WConfig, password)
	if err != nil {
		return err
	}

	proxy, e := pipeProxy.NewProxy(locAddr, w, t2s)
	if e != nil {
		return e
	}
	_instance = proxy
	return nil
}

func Proxying() {
	if _instance == nil {
		return
	}
	_instance.Proxying()
	_instance = nil
}

func StopVpn() {
	if _instance != nil {
		_instance.Done <- fmt.Errorf("user close this")
		_instance = nil
	}
}

func InputPacket(data []byte) error {

	if _instance == nil {
		return fmt.Errorf("tun isn't initilized ")
	}

	_instance.TunSrc.InputPacket(data)

	return nil
}

func VerifyAccount(addr, cipher, password string) bool {
	if _, err := account.AccFromString(addr, cipher, password); err != nil {
		fmt.Println("Valid Account:", err)
		return false
	}
	return true
}

func CreateAccount(password string) string {

	key, err := account.GenerateKey(password)
	if err != nil {
		return ""
	}
	address := key.ToNodeId().String()
	cipherTxt := base58.Encode(key.LockedKey)

	return address + Separator + cipherTxt
}

func IsProtonAddress(address string) bool {
	return account.ID(address).IsValid()
}

func LoadEthAddrByProtonAddr(protonAddr string) string {
	return ethereum.CheckProtonAddr(protonAddr)
}

func EthBindings(ETHAddr string) string {
	ethB, protonB, no := ethereum.BalanceOfEthAddr(ETHAddr)
	if ethB == nil {
		return ""
	}

	return fmt.Sprintf("%f"+Separator+"%f"+Separator+"%d",
		ethereum.ConvertByDecimal(ethB),
		ethereum.ConvertByDecimal(protonB),
		no)
}

func CreateEthAccount(password, directory string) string {
	return mobile.CreateEthAccount(password, directory)
}

func VerifyEthAccount(cipherTxt, pwd string) bool {
	return ethereum.VerifyEthAccount(cipherTxt, pwd)
}

func BindProtonAddress(protonAddr, cipherKey, password string) string {
	tx, err := ethereum.BindProtonAddr(protonAddr, cipherKey, password)
	if err != nil {
		fmt.Printf("\nBind proton addr(%s) err:%s", protonAddr, err)
		return err.Error()
	}
	return tx
}
func UnbindProtonAddress(protonAddr, cipherKey, password string) string {
	tx, err := ethereum.UnbindProtonAddr(protonAddr, cipherKey, password)
	if err != nil {
		fmt.Printf("\nBind proton addr(%s) err:%s", protonAddr, err)
		return err.Error()
	}
	return tx
}

func SetGlobalModel(global bool) {
	tun2Pipe.ByPassInst().ChangeGlobalModel(global)
}
func IsGlobalMode() bool {
	return tun2Pipe.ByPassInst().IsGlobalModel
}

func ReloadSeedNodes(url, path string) bool {
	nodes := pipeProxy.LoadFromServer(url)
	if e := ioutil.WriteFile(path, []byte(strings.Join(nodes, "\n")), 0644); e != nil {
		println("create boot nodes file failed:", path, e)
		return false
	}
	return true
}

func SetReturnModel(ret bool) {
	tun2Pipe.ByPassInst().IsReturnModel = ret
}
func IsReturnMode() bool {
	return tun2Pipe.ByPassInst().IsReturnModel
}
