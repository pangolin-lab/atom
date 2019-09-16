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

func (vp *VpnProxy) Accepting(err chan error, fetcher TargetFetcher, protocol payment.PacketPaymentProtocol) {

	fmt.Println("Proxy starting......")
	defer vp.Close()

	for {
		select {
		case e := <-err:
			fmt.Println("error from outside:", e)
			return
		default:

		}
		c, e := vp.conn.Accept()
		if e != nil {
			err <- e
			return
		}
		go vp.newPipeTask(c, fetcher, protocol)
	}
}

func (vp *VpnProxy) Close() {
	vp.conn.Close()
}
