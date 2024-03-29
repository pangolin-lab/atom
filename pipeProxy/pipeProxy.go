package pipeProxy

import (
	"fmt"
	"github.com/pangolin-lab/atom/wallet"
	"net"
	"strconv"
)

type PipeProxy struct {
	*net.TCPListener
	Done   chan error
	Wallet *wallet.Wallet
	TunSrc Tun2Pipe
}

func NewProxy(addr string, w *wallet.Wallet, t Tun2Pipe) (*PipeProxy, error) {
	l, e := net.Listen("tcp", addr)
	if e != nil {
		return nil, e
	}
	ap := &PipeProxy{
		TCPListener: l.(*net.TCPListener),
		Wallet:      w,
		TunSrc:      t,
		Done:        make(chan error),
	}
	return ap, nil
}

func (pp *PipeProxy) Proxying() {

	go pp.TunSrc.Proxying(pp.Done)

	go pp.Wallet.Running(pp.Done)

	go pp.Accepting(pp.Done)

	select {
	case err := <-pp.Done:
		fmt.Printf("PipeProxy exit for:%s", err.Error())
	}

	pp.Finish()
}

func (pp *PipeProxy) Accepting(done chan error) {

	fmt.Println("Proxy start working at:", pp.Addr().String())
	defer fmt.Println("Proxy exit......")

	for {
		conn, err := pp.Accept()
		if err != nil {
			fmt.Printf("\nFinish to proxy system request :%s", err)
			done <- err
			return
		}

		conn.(*net.TCPConn).SetKeepAlive(true)

		go pp.consume(conn)

		select {
		case err := <-done:
			fmt.Printf("\nProxy closed by out controller:%s", err.Error())
		default:
		}
	}
}

func (pp *PipeProxy) consume(conn net.Conn) {
	defer conn.Close()

	tgtAddr := pp.TunSrc.GetTarget(conn)

	if len(tgtAddr) < 10 {
		fmt.Println("\nNo such connection's target address:->", conn.RemoteAddr().String())
		return
	}
	fmt.Println("\n Proxying target address:", tgtAddr)

	//TODO::match PAC file in ios or android logic
	pipe := pp.Wallet.SetupPipe(conn, tgtAddr)
	if nil == pipe {
		fmt.Println("Create pipe failed:", tgtAddr)
		return
	}

	pipe.PullDataFromServer()

	rAddr := conn.RemoteAddr().String()
	_, port, _ := net.SplitHostPort(rAddr)
	keyPort, _ := strconv.Atoi(port)
	pp.TunSrc.RemoveFromSession(keyPort)

	//TODO::need to make sure is this ok
	fmt.Printf("\n\nPipe(%s) for(%s) is closing", rAddr, tgtAddr)
}

func (pp *PipeProxy) Finish() {

	if pp.TCPListener != nil {
		pp.TCPListener.Close()
		pp.TCPListener = nil
	}

	if pp.Wallet != nil {
		pp.Wallet.Finish()
	}

	if pp.TunSrc != nil {
		pp.TunSrc.Finish()
	}
}
