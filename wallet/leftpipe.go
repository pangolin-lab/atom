package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/proton-lab/proton-node/network"
	"github.com/proton-lab/proton-node/service/rpcMsg"
	"io"
	"net"
	"time"
)

type LeftPipe struct {
	reportChan  chan int
	target      string
	requestBuf  []byte
	responseBuf []byte
	proxyConn   net.Conn
	consume     *network.PipeConn
}

func NewPipe(l net.Conn, r *network.PipeConn, rc chan int, tgt string) *LeftPipe {
	return &LeftPipe{
		target:      tgt,
		requestBuf:  make([]byte, network.BuffSize),
		responseBuf: make([]byte, network.BuffSize),
		proxyConn:   l,
		consume:     r,
		reportChan:  rc,
	}
}

func (p *LeftPipe) collectRequest() {

	defer p.expire()
	defer fmt.Printf("collect system proxy conn for(%s) exit......", p.target)

	for {
		nr, err := p.proxyConn.Read(p.requestBuf)
		if nr > 0 {
			if nw, errW := p.consume.WriteCryptData(p.requestBuf[:nr]); errW != nil {
				fmt.Printf("\n forward system proxy err:%d, %v", nw, errW)
				return
			}
		}
		if err != nil {
			fmt.Printf("\n collet data for(%s) from client err:%v", p.target, err)
			return
		}
	}
}

func (p *LeftPipe) PullDataFromServer() {
	defer p.expire()
	defer fmt.Printf("\n consume conn for(%s) exit......", p.target)

	for {
		n, err := p.consume.ReadCryptData(p.responseBuf)
		p.reportChan <- n
		if n > 0 {
			if nw, errW := p.proxyConn.Write(p.responseBuf[:n]); errW != nil {
				fmt.Printf("\n Wallet Left pipe write data to system proxy err:%d, %v\n", nw, errW)
				return
			}
		}

		if err != nil {
			if err != io.EOF {
				fmt.Printf("\npull data from server:%v", err)
			}
			return
		}
	}
}

func (p *LeftPipe) expire() {
	p.consume.SetDeadline(time.Now())
	p.proxyConn.SetDeadline(time.Now())
}

func (p *LeftPipe) String() string {
	return fmt.Sprintf("%s<->%s for (%s)",
		p.proxyConn.RemoteAddr().String(),
		p.consume.RemoteAddr().String(), p.target)
}

func (w *Wallet) SetupPipe(lConn net.Conn, tgtAddr string) *LeftPipe {
	jsonConn, err := w.connectSockServer()
	if err != nil {
		fmt.Printf("\nConnet to socks server err:%v\n", err)
		return nil
	}

	if err := w.pipeHandshake(jsonConn, tgtAddr); err != nil {
		fmt.Printf("\nForward (%s) to server err:%v\n", tgtAddr, err)
		return nil
	}

	consumeConn := network.NewConsumerConn(jsonConn.Conn, w.aesKey)
	if consumeConn == nil {
		return nil
	}

	pipe := NewPipe(lConn, consumeConn, w.token, tgtAddr)

	fmt.Printf("\nNew pipe:%s ", pipe.String())

	go pipe.collectRequest()

	return pipe
}

func (w *Wallet) connectSockServer() (*network.JsonConn, error) {

	conn, err := w.getOuterConn(w.minerNetAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to (%s) access point server (%s):->", w.minerNetAddr, err)
	}
	conn.(*net.TCPConn).SetKeepAlive(true)
	return &network.JsonConn{Conn: conn}, nil
}

func (w *Wallet) pipeHandshake(conn *network.JsonConn, target string) error {

	reqData := &rpcMsg.PipeReqData{
		Addr:   w.acc.Address.String(),
		Target: target,
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("marshal hand shake data err:%v", err)
	}

	sig := w.acc.Sign(data)

	hs := &rpcMsg.YPHandShake{
		CmdType: rpcMsg.CmdPipe,
		Sig:     sig,
		Pipe:    reqData,
	}

	if err := conn.WriteJsonMsg(hs); err != nil {
		return fmt.Errorf("write hand shake data err:%v", err)

	}
	ack := &network.ProtonACK{}
	if err := conn.ReadJsonMsg(ack); err != nil {
		return fmt.Errorf("failed to read miner's response :->%v", err)
	}

	if !ack.Success {
		return fmt.Errorf("hand shake to miner err:%s", ack.Message)
	}

	return nil
}
