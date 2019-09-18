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
	conn    *network.JsonConn
	miner   *utils.PeerID
	accBook *accountBook
}

type accountBook struct {
	sync.RWMutex
	signal   chan int
	Counter  int
	Nonce    int
	UnSettle int64
}

const (
	receiptDBKey = "_receipt_local_cached_key_"
)

func (pw *PacketWallet) checkAccBook(poolAddr string, sig []byte, receipt *core.LatestReceipt) (*accountBook, error) {
	key := fmt.Sprintf("%s%s%s", receipt.UserAddr, receiptDBKey, poolAddr)
	if pw.sWallet.MainAddr != receipt.UserAddr || poolAddr != receipt.PoolAddr {
		return nil, fmt.Errorf("[PacketWallet-checkAccBook]  this[%s] is not my receipt", key)
	}

	accBook := &accountBook{signal: make(chan int, 100)}
	if err := utils.GetObj(pw.database, []byte(key), accBook); err != nil {
		fmt.Println("[PacketWallet-checkAccBook]  no cached receipt", err)
	}

	if accBook.Nonce == receipt.Nonce && accBook.UnSettle == receipt.UnSettle {
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
	pw.accBook.signal <- n
}

func (pw *PacketWallet) WriteCount(n int) {
	//Nothing to do
}

func (pw *PacketWallet) SynAccountBook() {

}

func (pw *PacketWallet) createChan(pool *ethereum.PoolDetail) (*Chanel, error) {

	conn, err := pw.connectToMiner(pool.Seeds)
	if err != nil {
		return nil, err
	}

	bootInfo, err := pw.handshake(conn)
	if err != nil {
		return nil, err
	}

	miner, err := pw.randomMiner(bootInfo.MinerIDs)
	if err != nil {
		return nil, err
	}

	accBook, err := pw.checkAccBook(pool.MainAddr, bootInfo.Sig, bootInfo.LatestReceipt)
	if err != nil {
		return nil, err
	}

	c := &Chanel{
		miner:   miner,
		conn:    conn,
		accBook: accBook,
	}
	return c, nil
}

func (pw *PacketWallet) monitor() {

	for {
		select {
		case n := <-pw.accBook.signal:
			pw.accBook.Counter += n
			if pw.accBook.Counter >= RechargeThreadHold {

			}

		case err := <-pw.errCh:
			pw.CloseChannel()
			fmt.Println("[PPP] monitor exit", err)
			return
		}
	}

}
