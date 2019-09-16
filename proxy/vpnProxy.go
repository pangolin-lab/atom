package proxy

import (
	"fmt"
	"github.com/pangolin-lab/atom/payment"
	"github.com/pangolin-lab/atom/utils"
	"net"
)

type VpnProxy struct {
	conn  net.Listener
	saver utils.ConnSaver
}

func NewProxyService(localSerAddr string, s utils.ConnSaver) (*VpnProxy, error) {

	c, e := net.Listen("tcp", localSerAddr)
	if e != nil {
		return nil, e
	}

	vp := &VpnProxy{
		conn:  c,
		saver: s,
	}
	return vp, nil
}

type TargetFetcher func(conn net.Conn) string

func (vp *VpnProxy) Accepting(result chan error, fetcher TargetFetcher, protocol payment.PacketPaymentProtocol) {

	fmt.Println("Proxy starting......")
	for {
		c, e := vp.conn.Accept()
		if e != nil {
			result <- e
			return
		}
		go vp.newPipeTask(c, fetcher, protocol)
	}
}

func (vp *VpnProxy) Close() {
	vp.conn.Close()
}
