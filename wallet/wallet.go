package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/proton-lab/proton-node/account"
	"github.com/proton-lab/proton-node/network"
	"github.com/proton-lab/proton-node/service/rpcMsg"
	"golang.org/x/crypto/ed25519"
	"log"
	"net"
	"sync"
	"syscall"
	"time"
)

type WConfig struct {
	BCAddr     string
	Cipher     string
	SettingUrl string
	ServerId   *ServeNodeId
	Saver      func(fd uintptr)
}

const MaxLocalConn = 1 << 10
const PipeDialTimeOut = time.Second * 2
const RechargeTimeInterval = time.Minute * 5

func (c *WConfig) String() string {
	return fmt.Sprintf("\n++++++++++++++++++++++++++++++++++++++++++++++++++++\n"+
		"+\t BCAddr:%s\n"+
		"+\t Ciphere:%s\n"+
		"+\tSettingUrl:%s\n"+
		"+\tServerId:%s\n"+
		"++++++++++++++++++++++++++++++++++++++++++++++++++++\n",
		c.BCAddr,
		c.Cipher,
		c.SettingUrl,
		c.ServerId.String())
}

type PacketBucket struct {
	sync.RWMutex
	token  chan int
	unpaid int
	total  int
}

type Wallet struct {
	*PacketBucket
	acc          *account.Account
	sysSaver     func(fd uintptr)
	payConn      *network.JsonConn
	aesKey       account.PipeCryptKey
	minerID      account.ID
	minerAddr    []byte
	minerNetAddr string
}

func NewWallet(conf *WConfig, password string) (*Wallet, error) {

	acc, err := account.AccFromString(conf.BCAddr, conf.Cipher, password)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n Unlock client success:%s Selected miner id:%s",
		conf.BCAddr, conf.ServerId.String())

	w := &Wallet{
		acc:          acc,
		minerID:      conf.ServerId.ID,
		sysSaver:     conf.Saver,
		minerNetAddr: conf.ServerId.TONetAddr(),
		PacketBucket: &PacketBucket{
			token: make(chan int, MaxLocalConn),
		},
	}
	w.minerAddr = make([]byte, len(w.minerID))
	copy(w.minerAddr, []byte(w.minerID))
	//TODO:: to be checked
	if err := w.acc.Key.GenerateAesKey(&w.aesKey, w.minerID.ToPubKey()); err != nil {
		return nil, err
	}

	if err := w.createRechargeChannel(); err != nil {
		log.Println("Create payment channel err:", err)
		return nil, err
	}

	fmt.Printf("\nCreate payment channel success:%s", w.ToString())

	return w, nil
}

func (w *Wallet) createRechargeChannel() error {
	fmt.Printf("\ncreatePayChannel Wallet socks ID addr:%s ", w.minerNetAddr)
	conn, err := w.getOuterConn(w.minerNetAddr)
	if err != nil {
		return err
	}

	sig := ed25519.Sign(w.acc.Key.PriKey, []byte(w.acc.Address))
	hs := &rpcMsg.YPHandShake{
		CmdType:  rpcMsg.CmdRecharge,
		Sig:      sig,
		UserAddr: w.acc.Address.String(),
	}

	jsonConn := &network.JsonConn{Conn: conn}
	if err := jsonConn.Syn(hs); err != nil {
		return err
	}

	w.payConn = jsonConn

	return nil
}

func (w *Wallet) Finish() {
	w.payConn.Close()
}

func (w *Wallet) getOuterConn(addr string) (net.Conn, error) {
	d := &net.Dialer{
		Timeout: PipeDialTimeOut,
		Control: func(network, address string, c syscall.RawConn) error {
			if w.sysSaver != nil {
				return c.Control(w.sysSaver)
			}
			return nil
		},
	}

	return d.Dial("tcp", addr)
}

func (w *Wallet) ToString() string {
	return fmt.Sprintf("\n++++++++++++++++++++++++++++++++++++++++++++++++++++\n"+
		"+\t account:%s\n"+
		"+\t minerID:%s\n"+
		"+\t Address:%s\n"+
		"++++++++++++++++++++++++++++++++++++++++++++++++++++\n",
		w.acc.Address,
		string(w.minerID),
		w.minerNetAddr)
}

func (w *Wallet) Running(done chan error) {

	for {
		select {
		case err := <-done:
			fmt.Printf("\nwallet closed by out controller:%s", err.Error())
		case no := <-w.token:
			if err := w.chargeUP(no); err != nil {
				fmt.Printf("\n Recharge failed:%s maybe I'll be cut off!", err.Error())
				done <- err
			}

		case <-time.After(RechargeTimeInterval):
			if err := w.timerRecharge(); err != nil {
				fmt.Printf("\n Timer recharge failed:%s maybe I'll be cut off!", err.Error())
				done <- err
			}
		}
	}
}

func (w *Wallet) timerRecharge() error {
	w.Lock()
	defer w.Unlock()

	fmt.Printf("\n  time to recharge report unpaid:%d", w.unpaid)
	if w.unpaid < rpcMsg.MinRechargeSize {
		return nil
	}

	if err := w.recharge(w.unpaid); err != nil {
		return err
	}

	w.unpaid = 0
	return nil
}

func (w *Wallet) chargeUP(no int) error {
	w.Lock()
	defer w.Unlock()
	w.unpaid += no

	fmt.Printf("\n  usage report unpaid:%d, this time:%d", w.unpaid, no)

	if w.unpaid < rpcMsg.RechargeUnit {
		return nil
	}

	if err := w.recharge(w.unpaid); err != nil {
		return err
	}

	w.unpaid = 0
	return nil
}

func CreatePayBill(user, miner string, usage int, priKey ed25519.PrivateKey) (*rpcMsg.UserCreditPay, error) {
	pay := &rpcMsg.CreditPayment{
		UserAddr:    user,
		MinerAddr:   miner,
		PacketUsage: usage,
		PayTime:     time.Now(),
	}

	data, err := json.Marshal(pay)
	if err != nil {
		return nil, err
	}
	sig := ed25519.Sign(priKey, data)

	return &rpcMsg.UserCreditPay{
		UserSig:       sig,
		CreditPayment: pay,
	}, nil
}

func (w *Wallet) recharge(no int) error {

	minerAddr := string(w.minerAddr)
	bill, err := CreatePayBill(string(w.acc.Address), minerAddr, no, w.acc.Key.PriKey)
	if err != nil {
		return err
	}

	fmt.Printf("Create new packet bill:%s for miner:%s", minerAddr, bill.String())

	if err := w.payConn.Syn(bill); err != nil {
		fmt.Printf("\nwallet write back bill msg err:%v", err)
		return err
	}

	fmt.Printf("recharge success:%d", no)
	return nil
}
