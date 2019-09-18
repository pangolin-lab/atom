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

type PacketPaymentProtocol interface {
	WalletAddr() (string, string)
	OpenPayChannel(errCh chan error, pool *ethereum.PoolDetail, auth string) error
	SetupAesConn(string) (account.CryptConn, error)
	IsPayChannelOpen(poolAddr string) bool
}

const (
	RechargeThreadHold = 1 << 12 //4M
	AccBookKeyJoin     = "@"
)

type SafeWallet struct {
	MainAddr  string
	SubAddr   string
	cipherTxt []byte
}

func initWallet(wPath string) (*SafeWallet, error) {
	data, err := ioutil.ReadFile(wPath)
	if err != nil {
		return nil, err
	}

	mAddr, sAddr, err := account.ParseWalletAddr(data)
	if err != nil {
		return nil, err
	}

	sw := &SafeWallet{
		MainAddr:  mAddr,
		SubAddr:   sAddr,
		cipherTxt: data,
	}
	fmt.Println("[InitProtocol] wallet initialization success......")
	return sw, nil
}

type PacketWallet struct {
	sWallet  *SafeWallet
	database *leveldb.DB
	wallet   account.Wallet
	pool     *ethereum.PoolDetail
	errCh    chan error
	*Chanel
}

func InitProtocol(wPath, rPath string) (PacketPaymentProtocol, error) {

	opts := opt.Options{
		Strict:      opt.DefaultStrict,
		Compression: opt.NoCompression,
		Filter:      filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(rPath, &opts)
	if err != nil {
		return nil, err
	}
	fmt.Println("[InitProtocol] open ppp database success......")

	sw, err := initWallet(wPath)
	if err != nil {
		fmt.Println("[InitProtocol]  empty wallet warning:", err)
		sw = &SafeWallet{}
	}

	//TODO::sync all packet balance from ethereum block chain contract
	pw := &PacketWallet{
		sWallet:  sw,
		database: db,
	}

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
			MainAddr: pw.sWallet.MainAddr,
			SubAddr:  pw.sWallet.SubAddr,
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
	pw.conn.Close()

	if pw.Chanel != nil {
		pw.synAccountBook()
	}

	pw.Chanel = nil
}

func (pw *PacketWallet) openWallet(auth string) error {
	if pw.sWallet.cipherTxt == nil {
		return fmt.Errorf("wallet data not found")
	}

	w, err := account.DecryptWallet(pw.sWallet.cipherTxt, auth)
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
	return pw.Chanel != nil
}

func (pw *PacketWallet) isWalletOpen() bool {
	return pw.wallet != nil && pw.wallet.IsOpen()
}

func (pw *PacketWallet) tryReopen() error {

	if !pw.wallet.IsOpen() {
		return fmt.Errorf("wallet has closed")
	}

	c, err := pw.createChan(pw.pool)
	if err != nil {
		return err
	}
	pw.Chanel = c
	go pw.monitor()
	return nil
}
