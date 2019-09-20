package payment

import (
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/utils"
	"github.com/pangolink/go-node/network"
	"github.com/pangolink/miner-pool/account"
	"github.com/pangolink/miner-pool/core"
	"sync"
)

type accountBook struct {
	EthBalance int64 `json:"eth"`
	LinBalance int64 `json:"token"`
	Counter    int   `json:"counter"`
	InRecharge int   `json:"charging"`
	Nonce      int   `json:"nonce"`
	UnClaimed  int64 `json:"unclaimed"`
}
type Accountant struct {
	sync.RWMutex
	signal    chan struct{}
	MainAddr  string `json:"mainAddress"`
	SubAddr   string `json:"subAddress"`
	CipherTxt []byte `json:"cipher"`

	*accountBook
}

type payChannel struct {
	pool  *ethereum.PoolDetail
	conn  *network.JsonConn
	miner *utils.PeerID
}

func (ac *Accountant) cacheKey() []byte {
	return []byte(fmt.Sprintf("%s%s%s", AccBookDataBaseKey, ac.MainAddr, ac.SubAddr))
}

func (ac *Accountant) loadAccBook(db *leveldb.DB) error {
	ac.Lock()
	defer ac.Unlock()

	key := ac.cacheKey()
	ok, err := db.Has(key, nil)
	if err != nil {
		return err
	}
	ac.accountBook = &accountBook{}
	if !ok {
		return nil
	}

	return utils.GetObj(db, key, ac.accountBook)
}

func (ac *Accountant) synBalance(db *leveldb.DB, cb SystemActionCallBack) {
	if ac.MainAddr == "" {
		return
	}

	ac.Lock()
	ac.LinBalance, ac.EthBalance = ethereum.TokenBalance(ac.MainAddr)
	ac.Unlock()

	ac.cacheAccBook(db)
	if cb != nil {
		cb.WalletBalanceSynced()
	}
}

func (ac *Accountant) cacheAccBook(db *leveldb.DB) {
	ac.RLock()
	defer ac.RUnlock()
	if err := utils.SaveObj(db, ac.cacheKey(), ac.accountBook); err != nil {
		fmt.Println("[PacketWallet-synBalance]  save cached err:", err)
	}
}

func (ac *Accountant) String() string {
	str := fmt.Sprintf("\n++++++++++++++++++++++++++++++++++++++++++++++++++++"+
		"+main address:\t%s"+
		"+sub address:\t%s"+
		"+CipherTxt:\t%s"+
		"+eth:\t%d"+
		"+token:\t%d"+
		"+Counter:\t%d"+
		"+InRecharge:\t%d"+
		"+Nonce:\t%d"+
		"+UnClaimed:\t%d"+
		"\n++++++++++++++++++++++++++++++++++++++++++++++++++++",
		ac.MainAddr,
		ac.SubAddr,
		ac.CipherTxt,
		ac.EthBalance,
		ac.LinBalance,
		ac.Counter,
		ac.InRecharge,
		ac.Nonce,
		ac.UnClaimed)
	return str
}

func (pw *PacketWallet) recharge(payee, payer string, w account.Wallet) error {
	req, err := pw.accBook.createRechargeReq(payee, payer, pw.payChan.miner.ID.String(), w)
	if err != nil {
		return err
	}

	if err := pw.payChan.conn.WriteJsonMsg(req); err != nil {
		return err
	}

	ack := &core.PayChanAck{}
	if err := pw.payChan.conn.ReadJsonMsg(ack); err != nil {
		return err
	}
	if !ack.Success {
		return fmt.Errorf("[PPP-recharge] err:%s", ack.ErrMsg)
	}

	if ack.Receipt.NextAction == core.PayResultRefresh {
		//TODO:: reload data from ethereum contract and recharge again
		return nil
	}

	pw.accBook.refresh(req.Recharge)
	return nil
}

func (ac *Accountant) ReadCount(n int) {
	ac.Lock()
	defer ac.Unlock()

	ac.Counter += n
	if ac.Counter >= RechargeThreadHold {
		ac.InRecharge += ac.Counter
		ac.Counter = 0
		ac.signal <- struct{}{}
	}
}

func (ac *Accountant) createRechargeReq(payee, payer, miner string, w account.Wallet) (*core.PayChanReq, error) {
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
			Balance:   0, //TODO::
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

func (ac *Accountant) refresh(req *core.PacketRecharge) {
	ac.Lock()
	defer ac.Unlock()

	ac.InRecharge -= req.Usage
	ac.UnClaimed += int64(req.Usage)
}

func (ac *Accountant) WriteCount(n int) {
	//Nothing to do
}

func (pw *PacketWallet) createChan(pool *ethereum.PoolDetail) (*payChannel, error) {

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

	c := &payChannel{
		miner: miner,
		conn:  conn,
		pool:  pool,
	}
	return c, nil
}

func (pw *PacketWallet) monitor() {
	defer pw.CloseChannel()
	defer fmt.Println("[PPP] Channel closed------>")

	for {
		select {
		case <-pw.accBook.signal:
			if err := pw.recharge(pw.accBook.MainAddr, pw.payChan.pool.MainAddr, pw.wallet); err != nil {
				pw.errCh <- err
				fmt.Println("[PPP] Count error:", err)
				return
			}

			pw.accBook.cacheAccBook(pw.database)
		case err := <-pw.errCh:
			fmt.Println("[PPP] monitor exit:", err)
			return
		}
	}
}

func (pw *PacketWallet) checkAccBook(poolAddr string, sig []byte, receipt *core.LatestReceipt) (*Accountant, error) {
	if pw.accBook.MainAddr != receipt.UserAddr || poolAddr != receipt.PoolAddr {
		return nil, fmt.Errorf("[PacketWallet-checkAccBook] is not my receipt")
	}

	accBook := pw.accBook

	if accBook.Nonce == receipt.Nonce &&
		accBook.UnClaimed == receipt.UnClaimed {

		return accBook, nil
	}

	if ok := pw.wallet.VerifySig(sig, receipt); ok {
		return nil, fmt.Errorf("verify pool's receipt failed")
	}

	accBook.Nonce = receipt.Nonce
	accBook.UnClaimed = receipt.UnClaimed
	return accBook, nil
}
