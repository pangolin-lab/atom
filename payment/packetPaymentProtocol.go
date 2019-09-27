package payment

import (
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/utils"
	"github.com/pangolink/go-node/network"
	"github.com/pangolink/miner-pool/account"
	"github.com/pangolink/miner-pool/core"
	"io/ioutil"
	"sort"
	"sync"
	"time"
)

const (
	RechargeThreadHold = 1 << 12 //4M
	AccBookKeyJoin     = "@"
	AccBookDataBaseKey = "_acc_book_database_key_"
)

type SystemActionCallBack interface {
	WalletBalanceSynced()
}

type PacketPaymentProtocol interface {
	OpenPayChannel(errCh chan error, pool *ethereum.PoolDetail, auth string) error
	SetupAesConn(target string) (account.CryptConn, error)
	IsPayChannelOpen(poolAddr string) bool
	Finalized()
	AccBookInfo() *Accountant
	SyncWalletBalance()
	Wallet(auth string) (account.Wallet, error)
	NewWallet(auth, wp string) (account.Wallet, error)
}

type PacketWallet struct {
	database *leveldb.DB
	errCh    chan error
	callBack SystemActionCallBack

	accBook *Accountant
	wallet  account.Wallet
	payChan *payChannel
}

func initAcc(wPath string, db *leveldb.DB) (*Accountant, error) {

	if _, ok := utils.FileExists(wPath); !ok {
		fmt.Println("[initAcc]  FileExists doesn't exit:", wPath)
		return &Accountant{
			signal: make(chan struct{}),
		}, nil
	}

	data, err := ioutil.ReadFile(wPath)
	if err != nil {
		fmt.Println("[initAcc]  ioutil read file err:", err)
		return nil, err
	}

	mAddr, sAddr, err := account.ParseWalletAddr(data)
	if err != nil {
		fmt.Println("[initAcc]  ParseWalletAddr err:", err)
		return nil, err
	}

	ab := &Accountant{
		signal:    make(chan struct{}),
		MainAddr:  mAddr,
		SubAddr:   sAddr,
		CipherTxt: data,
	}

	if err := ab.loadAccBook(db); err != nil {
		fmt.Println("[initAcc] loadAccBook err:", err)
		return nil, err
	}
	fmt.Println("[initAcc] accountant initialization success......", ab.String())
	return ab, nil
}

func InitProtocol(wPath, rPath string, cb SystemActionCallBack) (PacketPaymentProtocol, error) {

	opts := opt.Options{
		Strict:      opt.DefaultStrict,
		Compression: opt.NoCompression,
		Filter:      filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(rPath, &opts)
	if err != nil {
		fmt.Println("leveldb open failed:", err)
		return nil, err
	}
	fmt.Println("[InitProtocol] open ppp database success......")

	ab, err := initAcc(wPath, db)
	if err != nil {
		fmt.Println("initAcc failed:", err)
		return nil, err
	}
	pw := &PacketWallet{
		database: db,
		errCh:    make(chan error, 10),
		callBack: cb,
		accBook:  ab,
	}
	go ab.synBalance(db, cb)

	fmt.Println("[InitProtocol] packet payment protocol success......")
	return pw, nil
}

func (pw *PacketWallet) connectToMiner(poolNodeId string) (*network.JsonConn, error) {
	peerId, e := utils.ConvertPID(poolNodeId)
	if e != nil {
		return nil, e
	}
	conn, err := utils.GetSavedConn(peerId.NetAddr)
	if err != nil {
		return nil, err
	}
	c := &network.JsonConn{Conn: conn}
	return c, nil
}

func (pw *PacketWallet) handshake(conn *network.JsonConn) (*core.ChanCreateAck, error) {

	req := &core.PayChanReq{
		MsgType: core.CreateReq,
		CreateReq: &core.ChanCreateReq{
			MainAddr: pw.accBook.MainAddr,
			SubAddr:  pw.accBook.SubAddr,
		},
	}
	req.Sig = pw.wallet.SignSub(req)
	if err := conn.WriteJsonMsg(req); err != nil {
		return nil, err
	}

	ack := &core.PayChanAck{}
	if err := conn.ReadJsonMsg(ack); err != nil {
		return nil, err
	}

	if ack.Success != true {
		return nil, fmt.Errorf("create new coin payee err:%s", ack.ErrMsg)
	}
	return ack.CreateRes, nil
}

func (pw *PacketWallet) CloseChannel() {

	if pw.payChan != nil {
		pw.accBook.cacheAccBook(pw.database)
		pw.payChan.conn.Close()
	}
	pw.payChan = nil
}

func (pw *PacketWallet) openWallet(auth string) error {
	if pw.accBook.CipherTxt == nil {
		return fmt.Errorf("wallet data not found")
	}

	w, err := account.DecryptWallet(pw.accBook.CipherTxt, auth)
	if err != nil {
		return err
	}

	pw.wallet = w
	return nil
}

func (pw *PacketWallet) randomMiner(minerIDs []string) (*utils.PeerID, error) {

	var waiter sync.WaitGroup
	s := make([]*utils.PeerID, 0)
	var locker sync.Mutex

	for _, id := range minerIDs {
		pid, err := utils.ConvertPID(id)
		if err != nil {
			continue
		}

		waiter.Add(1)
		go func() {
			defer waiter.Done()

			pid.TTL()
			fmt.Printf("\nserver(%s) is ok (%dms)\n", pid.IP, pid.Ping/time.Millisecond)
			locker.Lock()
			s = append(s, pid)
			locker.Unlock()
		}()
	}
	waiter.Wait()

	if len(s) == 0 {
		return nil, fmt.Errorf("[randomMiner] no valid miner node")
	}

	sort.Slice(s, func(i, j int) bool {
		return s[i].Ping < s[j].Ping
	})
	return s[0], nil
}

func (pw *PacketWallet) isChanOpen() bool {
	return pw.payChan != nil
}

func (pw *PacketWallet) isWalletOpen() bool {
	return pw.wallet != nil && pw.wallet.IsOpen()
}

func (pw *PacketWallet) tryReopen() error {

	if !pw.wallet.IsOpen() {
		return fmt.Errorf("wallet has closed")
	}

	c, err := pw.createChan(pw.payChan.pool)
	if err != nil {
		return err
	}
	pw.payChan = c
	go pw.monitor()
	return nil
}
