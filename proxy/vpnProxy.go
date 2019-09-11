package proxy

import (
	"fmt"
	"github.com/pangolin-lab/atom/microPay"
	"github.com/pangolin-lab/atom/utils"
	"net"
)

type VpnProxy struct {
	conn    net.Listener
	saver   utils.ConnSaver
	payChan microPay.PayChannel
	miner   *utils.PeerID
}

func NewProxyService(localSerAddr string, pc microPay.PayChannel, s utils.ConnSaver) (*VpnProxy, error) {

	c, e := net.Listen("tcp", localSerAddr)
	if e != nil {
		return nil, e
	}

	vp := &VpnProxy{
		conn:    c,
		saver:   s,
		payChan: pc,
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
