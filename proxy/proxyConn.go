package proxy

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

const PipeDialTimeOut = time.Second * 2

var connSaver ConnSaver = nil

type SocksConn struct {
	conn net.Conn
}

func (vp *VpnProxy) NewReqThread(c net.Conn) {
	defer c.Close()

	c.(*net.TCPConn).SetKeepAlive(true)

	tgtHost := GetTarget(c)
	fmt.Printf("\n New conn thread for target[%s]:", tgtHost)

	d := &net.Dialer{
		Timeout: PipeDialTimeOut,
		Control: func(network, address string, c syscall.RawConn) error {
			if connSaver != nil {
				return c.Control(connSaver)
			}
			return nil
		},
	}

	_, err := d.Dial("tcp", vp.miner)
	if err != nil {
		return
	}

}
