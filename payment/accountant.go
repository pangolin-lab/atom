package payment

import (
	"github.com/pangolin-lab/atom/utils"
	"github.com/pangolink/go-node/network"
	"sync"
)

type accountBook struct {
	sync.RWMutex
	Counter  int
	Nonce    int
	UnSettle int64
}

type PayeeInfo struct {
	PayeeAddr string
	SelMiner  *utils.PeerID
	conn      *network.JsonConn
}

func (pw *PacketWallet) ReadCount(n int) {

}

func (pw *PacketWallet) WriteCount(n int) {

}
