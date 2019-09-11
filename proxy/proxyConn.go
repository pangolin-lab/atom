package proxy

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

type SocksConn struct {
	conn net.Conn
}

func (vp *VpnProxy) NewReqThread(c net.Conn) {
	defer c.Close()

	c.(*net.TCPConn).SetKeepAlive(true)

	tgtHost := GetTarget(c)
	fmt.Printf("\n New conn thread for target[%s]:", tgtHost)
}
