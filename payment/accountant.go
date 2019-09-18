package payment

import (
	"fmt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/utils"
	"github.com/pangolink/go-node/network"
	"github.com/pangolink/miner-pool/account"
	"github.com/pangolink/miner-pool/core"
	"sync"
)

const (
	receiptDBKey = "_receipt_local_cached_key_"
)

/************************************************************************************
*Chanel*
************************************************************************************/
type Chanel struct {
	conn    *network.JsonConn
	miner   *utils.PeerID
	accBook *accountBook
}

func (ch *Chanel) recharge(payee, payer string, w account.Wallet) error {

	req, err := ch.accBook.createRechargeReq(payee, payer, ch.miner.ID.String(), w)
	if err != nil {
		return err
	}

	if err := ch.conn.WriteJsonMsg(req); err != nil {
		return err
	}

	ack := &core.PayChanAck{}
	if err := ch.conn.ReadJsonMsg(ack); err != nil {
		return err
	}
	if !ack.Success {
		return fmt.Errorf("[PPP-recharge] err:%s", ack.ErrMsg)
	}

	if ack.Receipt.NextAction == core.PayResultRefresh {
		//TODO:: reload data from ethereum contract and recharge again
		return nil
	}
	//
	//if req.Recharge.Usage != ack.Receipt.Recharged{
	//	return fmt.Errorf("[PPP-recharge] Cheating miner pool[%s], usage(%d-%d) not same",
	//		payee, req.Recharge.Usage, ack.Receipt.Recharged)
	//}

	ch.accBook.refresh(req.Recharge)
	return nil
}

/************************************************************************************
*accountBook*
************************************************************************************/

type accountBook struct {
	sync.RWMutex
	signal     chan struct{}
	Counter    int
	InRecharge int
	Nonce      int
	UnClaimed  int64
	Balance    int64
}

func (ac *accountBook) ReadCount(n int) {
	ac.Lock()
	defer ac.Unlock()

	ac.Counter += n
	if ac.Counter >= RechargeThreadHold {
		ac.InRecharge += ac.Counter
		ac.Counter = 0
		ac.signal <- struct{}{}
	}
}

func (ac *accountBook) createRechargeReq(payee, payer, miner string, w account.Wallet) (*core.PayChanReq, error) {
	ac.RLock()
	defer ac.RUnlock()

	req := &core.PayChanReq{
		MsgType: core.Recharge,
		Recharge: &core.PacketRecharge{
			Payee:     payee,
			Payer:     payer,
			Miner:     miner,
			Usage:     ac.InRecharge,
			Contract:  ethereum.Conf.MicroPaySys,
			UnClaimed: ac.UnClaimed,
			Balance:   ac.Balance,
			Nonce:     ac.Nonce,
		},
	}

	sig, err := w.Sign(req.Recharge)
	if err != nil {
		return nil, err
	}
	req.Sig = sig

	return req, nil
}

func (ac *accountBook) refresh(req *core.PacketRecharge) {
	ac.Lock()
	defer ac.Unlock()

	ac.InRecharge -= req.Usage
	ac.UnClaimed += int64(req.Usage)
}

func (ac *accountBook) WriteCount(n int) {
	//Nothing to do
}

/************************************************************************************
*PacketWallet*
************************************************************************************/

func (pw *PacketWallet) synAccountBook() {
	ab := pw.accBook
	ab.RLock()
	defer ab.RUnlock()

	key := fmt.Sprintf("%s%s%s", pw.sWallet.MainAddr, receiptDBKey, pw.pool.MainAddr)
	if err := utils.SaveObj(pw.database, []byte(key), ab); err != nil {
		fmt.Println("[PacketWallet-synAccountBook]  save cached err:", err)
	}
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
	defer pw.CloseChannel()
	defer fmt.Println("[PPP] Channel closed------>")

	for {
		select {
		case <-pw.accBook.signal:
			if err := pw.Chanel.recharge(pw.sWallet.MainAddr, pw.pool.MainAddr, pw.wallet); err != nil {
				pw.errCh <- err
				fmt.Println("[PPP] Count error:", err)
				return
			}

			pw.synAccountBook()
		case err := <-pw.errCh:
			fmt.Println("[PPP] monitor exit:", err)
			return
		}
	}
}

func (pw *PacketWallet) checkAccBook(poolAddr string, sig []byte, receipt *core.LatestReceipt) (*accountBook, error) {

	key := fmt.Sprintf("%s%s%s", receipt.UserAddr, receiptDBKey, poolAddr)
	if pw.sWallet.MainAddr != receipt.UserAddr || poolAddr != receipt.PoolAddr {
		return nil, fmt.Errorf("[PacketWallet-checkAccBook]  this[%s] is not my receipt", key)
	}

	accBook := &accountBook{signal: make(chan struct{})}
	if err := utils.GetObj(pw.database, []byte(key), accBook); err != nil {
		fmt.Println("[PacketWallet-checkAccBook]  no cached receipt", err)
	}

	if accBook.Nonce == receipt.Nonce &&
		accBook.UnClaimed == receipt.UnClaimed &&
		accBook.Balance == receipt.Balance {

		return accBook, nil
	}

	if ok := pw.wallet.VerifySig(sig, receipt); ok {
		return nil, fmt.Errorf("verify pool's receipt failed")
	}

	accBook.Nonce = receipt.Nonce
	accBook.UnClaimed = receipt.UnClaimed
	accBook.Balance = receipt.Balance

	return accBook, nil
}
