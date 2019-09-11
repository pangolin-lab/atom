package proxy

import (
	"fmt"
	"github.com/pangolin-lab/atom/utils"
	"github.com/pangolink/go-node/network"
	"net"
)

type Pipe struct {
	target      string
	requestBuf  []byte
	responseBuf []byte
	localConn   net.Conn
	remoteConn  *network.PipeConn
}

func (p *Pipe) Close() {
	p.remoteConn.Close()
	p.localConn.Close()
}

func (vp *VpnProxy) newProxyPipe(c net.Conn, fetcher TargetFetcher) {

	c.(*net.TCPConn).SetKeepAlive(true)
	pipe := &Pipe{
		localConn: c,
	}
	defer pipe.Close()

	tgtHost := fetcher(c)
	if len(tgtHost) < 2 {
		fmt.Printf("\n Invalid target[%s]:", tgtHost)
		return
	}
	pipe.target = tgtHost

	_, err := utils.GetSavedConn(vp.payChan.MinerNetAddr())
	if err != nil {
		fmt.Printf("\nConnect to miner failed:[%s]", err.Error())
		return
	}

}
