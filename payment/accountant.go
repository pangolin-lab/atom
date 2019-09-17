package payment

import (
	"fmt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/utils"
	"github.com/pangolink/go-node/network"
	"github.com/pangolink/miner-pool/core"
	"sync"
)

type Chanel struct {
	pool    *ethereum.PoolDetail
	conn    *network.JsonConn
	miner   *utils.PeerID
	accBook *accountBook
}

type accountBook struct {
	sync.RWMutex
	Counter  int
	Nonce    int
	UnSettle int64
}

const (
	receiptDBKey = "_receipt_local_cached_key_"
)

func (pw *PacketWallet) checkAccount(poolAddr string, sig []byte, receipt *core.LatestReceipt) (*accountBook, error) {
	key := fmt.Sprintf("%s%s%s", receipt.UserAddr, receiptDBKey, poolAddr)
	if pw.sWallet.MainAddr != receipt.UserAddr || poolAddr != receipt.PoolAddr {
		return nil, fmt.Errorf("[PacketWallet-checkAccount]  this[%s] is not my receipt", key)
	}

	accBook := &accountBook{}
	if err := utils.GetObj(pw.database, []byte(key), accBook); err != nil {
		fmt.Println("[PacketWallet-checkAccount]  no cached receipt", err)
	}

	if accBook.Nonce == receipt.Nonce && accBook.UnSettle != receipt.UnSettle {
		return accBook, nil
	}
	if ok := pw.wallet.VerifySig(sig, receipt); ok {
		return nil, fmt.Errorf("verify pool's receipt failed")
	}
	accBook.Nonce = receipt.Nonce
	accBook.UnSettle = receipt.UnSettle
	return accBook, nil
}

func (pw *PacketWallet) ReadCount(n int) {

}

func (pw *PacketWallet) WriteCount(n int) {

}

func (pw *PacketWallet) SynAccountBook() {

}
