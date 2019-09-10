package proxy

import (
	"fmt"
	"github.com/pangolink/miner-pool/account"
	"net"
)

type VpnProxy struct {
	conn   net.Listener
	saver  ConnSaver
	wallet account.Wallet
	miner  string
}

type ConnSaver func(fd uintptr)

func NewProxyService(addr, url string, w account.Wallet, s ConnSaver) (*VpnProxy, error) {

	c, e := net.Listen("tcp", addr)
	if e != nil {
		return nil, e
	}

	vp := &VpnProxy{
		conn:   c,
		saver:  s,
		wallet: w,
	}

	return vp, nil
}

func (vp *VpnProxy) Accepting(result chan string) {
	fmt.Println("Proxy starting......")

	for {
		c, e := vp.conn.Accept()
		if e != nil {
			result <- e.Error()
			return
		}

		go vp.NewReqThread(c)
	}
}
