package microPay

import "C"
import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"github.com/pangolin-lab/atom/utils"
	acc "github.com/pangolink/go-node/account"
	"github.com/pangolink/go-node/network"
	"github.com/pangolink/go-node/service/rpcMsg"
	"github.com/pangolink/miner-pool/account"
	"github.com/pangolink/miner-pool/core"
	"io"
	"sync"
)

const (
	RechargeThreadHold = 1 << 12 //4M
)

type PayChannel interface {
	SetupAesConn(tgt string) (*account.AesConn, error)
	Close()
}

type MinerNode struct {
	ID      *utils.PeerID
	NetAddr string
	AesKey  []byte
}

type micPayChan struct {
	sync.RWMutex
	rCounter int
	wallet   account.Wallet
	conn     *network.JsonConn
	miner    *MinerNode
}

func NewChannel(cipherTxt, auth, poolNode string) (PayChannel, error) {
	w, e := account.DecryptWallet([]byte(cipherTxt), auth)
	if e != nil {
		return nil, e
	}

	peerId, e := utils.ConvertPID(poolNode)
	if e != nil {
		return nil, e
	}
	conn, err := utils.GetSavedConn(peerId.NetAddr())
	if err != nil {
		return nil, err
	}
	c := &network.JsonConn{Conn: conn}

	minerId, err := handShake(w, c)
	if err != nil {
		return nil, err
	}
	miner, err := newMiner(minerId, w)
	if err != nil {
		return nil, err
	}

	m := &micPayChan{
		wallet: w,
		conn:   c,
		miner:  miner,
	}
	return m, nil
}

func handShake(w account.Wallet, conn *network.JsonConn) (string, error) {
	mAddr, sAddr := w.Address()
	syn := &core.PayChanSyn{
		MsgType: core.CreateReq,
		CreateReq: &core.ChanCreateReq{
			MainAddr: mAddr,
			SubAddr:  sAddr,
		},
	}
	sig, err := w.Sign(syn.CreateReq)
	if err != nil {
		return "", err
	}
	syn.Sig = sig

	if err := conn.WriteJsonMsg(syn); err != nil {
		return "", err
	}
	ack := &core.PayChanAck{}
	if err := conn.ReadJsonMsg(ack); err != nil {
		return "", err
	}

	if ack.Success != true {
		return "", fmt.Errorf("create payment channel err:%s", ack.ErrMsg)
	}
	return ack.MinerId, nil
}

func newMiner(minerId string, w account.Wallet) (*MinerNode, error) {

	mid, e := utils.ConvertPID(minerId)
	if e != nil {
		return nil, utils.ErrInvalidID
	}

	aesKey := new(acc.PipeCryptKey)
	if err := acc.GenerateAesKey(aesKey, mid.ID.ToPubKey(), w.CryptKey()); err != nil {
		return nil, err
	}
	m := &MinerNode{
		ID:      mid,
		NetAddr: mid.NetAddr(),
		AesKey:  make([]byte, len(aesKey)),
	}
	copy(m.AesKey, aesKey[:])
	return m, nil
}

func (mpc *micPayChan) SetupAesConn(target string) (*account.AesConn, error) {
	conn, err := utils.GetSavedConn(mpc.miner.NetAddr)
	if err != nil {
		fmt.Printf("\nConnect to miner failed:[%s]", err.Error())
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	io.ReadFull(rand.Reader, iv[:])

	//TODO:: iv, target, sub addr
	jsonConn := network.JsonConn{Conn: conn}
	_, subAddr := mpc.wallet.Address()
	req := rpcMsg.AesConnSetup{
		IV:          iv[:],
		Target:      target,
		UserSubAddr: subAddr,
	}

	req.Sig = mpc.wallet.SignSub(req)
	if err := jsonConn.Syn(req); err != nil {
		fmt.Println("Send salt to miner failed:", err)
		return nil, err
	}

	block, err := aes.NewCipher(mpc.miner.AesKey)
	if err != nil {
		return nil, err
	}

	ac := &account.AesConn{
		Conn:    conn,
		Counter: mpc,
		Encoder: cipher.NewCFBEncrypter(block, iv),
		Decoder: cipher.NewCFBDecrypter(block, iv),
	}
	return ac, nil
}

func (mpc *micPayChan) Close() {
	mpc.conn.Close()
	mpc.wallet.Close()
}

func (mpc *micPayChan) ReadCount(n int) {
	mpc.Lock()
	defer mpc.Unlock()
	mpc.rCounter += n
	if mpc.rCounter >= RechargeThreadHold {
		go mpc.microPay()
	}
}

func (mpc *micPayChan) WriteCount(n int) {
	//Stub don't care
}

func (mpc *micPayChan) microPay() {
	receipt := core.MicPayReceipt{}
	mpc.conn.WriteJsonMsg(receipt)
}
