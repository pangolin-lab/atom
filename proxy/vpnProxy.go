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

type TargetFetcher func(conn net.Conn) string

func (vp *VpnProxy) Accepting(result chan string, fetcher TargetFetcher) {

	fmt.Println("Proxy starting......")

	for {
		c, e := vp.conn.Accept()
		if e != nil {
			result <- e.Error()
			return
		}
		go vp.newPipeTask(c, fetcher)
	}
}
