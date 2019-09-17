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
	OpenPayChannel(errCh chan error, poolId *ethereum.PoolDetail, auth string) error
	SetupAesConn(string) (account.CryptConn, error)
	IsPayChannelOpen(poolAddr string) bool
}

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
	return sw, nil
}

type PacketWallet struct {
	sWallet  *SafeWallet
	database *leveldb.DB
	wallet   account.Wallet
	*Chanel
}

func InitProtocol(wPath, rPath string) (PacketPaymentProtocol, error) {

	opts := opt.Options{
		ErrorIfExist: true,
		Strict:       opt.DefaultStrict,
		Compression:  opt.NoCompression,
		Filter:       filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(rPath, &opts)
	if err != nil {
		return nil, err
	}

	sw, err := initWallet(wPath)
	if err != nil {
		fmt.Println("[PPP] InitProtocol initWallet err:", err)
		sw = &SafeWallet{}
	}

	pw := &PacketWallet{
		sWallet:  sw,
		database: db,
	}
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

func (pw *PacketWallet) monitor(errors chan error) {

}

func (pw *PacketWallet) CloseChannel() {
	pw.conn.Close()

	if pw.Chanel != nil {
		pw.SynAccountBook()
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
