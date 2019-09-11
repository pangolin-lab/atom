package utils

import (
	"net"
	"syscall"
	"time"
)

type ConnSaver func(fd uintptr)

const PipeDialTimeOut = time.Second * 2

var connSaver ConnSaver = nil

func GetSavedConn(rAddr string) (net.Conn, error) {
	d := &net.Dialer{
		Timeout: PipeDialTimeOut,
		Control: func(network, address string, c syscall.RawConn) error {
			if connSaver != nil {
				return c.Control(connSaver)
			}
			return nil
		},
	}

	return d.Dial("tcp", rAddr)
}
