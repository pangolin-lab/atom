package androidLib

import "C"
import (
	"fmt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/pipeProxy"
	"github.com/pangolin-lab/atom/tun2Pipe"
	"github.com/pangolin-lab/atom/wallet"
	"github.com/pangolink/miner-pool/common"
	"io/ioutil"
	"strings"
)

type VpnDelegate interface {
	tun2Pipe.VpnDelegate
	GetBootPath() string
}

var _instance *pipeProxy.PipeProxy = nil
var proxyConf = &pipeProxy.ProxyConfig{}

func InitVPN(addr, cipher, url, boot, IPs string, d VpnDelegate) error {
	ethereum.Conf = common.TestNet
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

func ReloadSeedNodes(url, path string) bool {
	nodes := pipeProxy.LoadFromServer(url)
	if e := ioutil.WriteFile(path, []byte(strings.Join(nodes, "\n")), 0644); e != nil {
		println("create boot nodes file failed:", path, e)
		return false
	}
	return true
}
