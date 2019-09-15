package proxy

import (
	"fmt"
	"github.com/pangolin-lab/atom/microPay"
	"github.com/pangolink/go-node/network"
	"github.com/pangolink/miner-pool/account"
	"io"
	"net"
)

type Pipe struct {
	target string
	lBuf   []byte
	rBuf   []byte
	lConn  net.Conn
	rConn  account.CryptConn
}

func (p *Pipe) Close() {
	p.rConn.Close()
	p.lConn.Close()
}

func (p *Pipe) transMinerData() {
	defer p.Close()
	defer fmt.Printf("\n trans miner data for[%s] thread exit:", p.target)

	for {
		n, err := p.rConn.ReadCryptData(p.rBuf)
		if n > 0 {
			if _, lErr := p.lConn.Write(p.rBuf[:n]); lErr != nil {
				fmt.Println(lErr)
				return
			}
		}
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			return
		}
	}
}

func (p *Pipe) transLocalData() {
	defer p.Close()
	defer fmt.Printf("\n trans local request data for[%s] thread exit:", p.target)

	for {
		n, err := p.lConn.Read(p.lBuf)
		if n > 0 {
			if _, rErr := p.rConn.WriteCryptData(p.lBuf[:n]); rErr != nil {
				fmt.Println(rErr)
				return
			}
		}

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			return
		}
	}
}

func (vp *VpnProxy) newPipeTask(c net.Conn, fetcher TargetFetcher, payChan microPay.PayChannel) {
	c.(*net.TCPConn).SetKeepAlive(true)

	tgtHost := fetcher(c)
	if len(tgtHost) < 2 {
		fmt.Printf("\n Invalid target[%s]:", tgtHost)
		return
	}

	aesConn, err := payChan.SetupAesConn(tgtHost)
	if err != nil {
		fmt.Printf("Create connection to miner for [%s] err:%s", tgtHost, err.Error())
		return
	}

	pipe := &Pipe{
		lConn:  c,
		target: tgtHost,
		rConn:  aesConn,
		lBuf:   make([]byte, network.BuffSize),
		rBuf:   make([]byte, network.BuffSize),
	}
	go pipe.transMinerData()
	pipe.transLocalData()
}
