package payment

import "C"
import (
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"github.com/pangolin-lab/atom/utils"
	acc "github.com/pangolink/go-node/account"
	"github.com/pangolink/go-node/network"
	"github.com/pangolink/go-node/service/rpcMsg"
	"github.com/pangolink/miner-pool/account"
	"github.com/pangolink/miner-pool/core"
	"io"
)

type PayChannel interface {
	SetupAesConn(tgt string) (account.CryptConn, error)
	Close()
	IsOpen() bool
}

type minerInfo struct {
	ID      *utils.PeerID
	NetAddr string
	AesKey  []byte
}

type micPayChan struct {
	accBook *AccBook
	wallet  account.Wallet
	conn    *network.JsonConn
	miner   *minerInfo
	pool    *utils.PeerID
}

func NewChannel(cipherTxt, auth, poolNode, accPath string) (PayChannel, error) {
	w, e := account.DecryptWallet([]byte(cipherTxt), auth)
	if e != nil {
		return nil, e
	}

	peerId, e := utils.ConvertPID(poolNode)
	if e != nil {
		return nil, e
	}
	conn, err := utils.GetSavedConn(peerId.NetAddr)
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

	mAddr, _ := w.Address()
	ab, err := loadAccBook(accPath, mAddr, poolNode)
	if err != nil {
		return nil, err
	}

	m := &micPayChan{
		wallet:  w,
		conn:    c,
		miner:   miner,
		pool:    peerId,
		accBook: ab,
	}
	return m, nil
}

func handShake(w account.Wallet, conn *network.JsonConn) (string, error) {
	mAddr, sAddr := w.Address()
	syn := &core.PayChanReq{
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
	return ack.CreateRes.MinerIDs[0], nil
}

func newMiner(minerId string, w account.Wallet) (*minerInfo, error) {

	mid, e := utils.ConvertPID(minerId)
	if e != nil {
		return nil, utils.ErrInvalidID
	}

	aesKey := new(acc.PipeCryptKey)
	if err := acc.GenerateAesKey(aesKey, mid.ID.ToPubKey(), w.CryptKey()); err != nil {
		return nil, err
	}
	m := &minerInfo{
		ID:      mid,
		NetAddr: mid.NetAddr,
		AesKey:  make([]byte, len(aesKey)),
	}
	copy(m.AesKey, aesKey[:])
	return m, nil
}

func (mpc *micPayChan) SetupAesConn(target string) (account.CryptConn, error) {
	conn, err := utils.GetSavedConn(mpc.miner.NetAddr)
	if err != nil {
		fmt.Printf("\nConnect to miner failed:[%s]", err.Error())
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	io.ReadFull(rand.Reader, iv[:])

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

	return account.NewAesConn(conn, mpc, mpc.miner.AesKey, iv)
}
func (mpc *micPayChan) IsOpen() bool {
	return mpc.wallet == nil
}

func (mpc *micPayChan) Close() {
	mpc.conn.Close()
	mpc.wallet.Close()
	mpc.wallet = nil
}

func (mpc *micPayChan) ReadCount(n int) {
	mpc.accBook.incrUsage(n)
	if mpc.accBook.Counter >= RechargeThreadHold {
		go mpc.microPay()
	}
}

func (mpc *micPayChan) WriteCount(n int) {
	//Stub don't care
}

//TODO::notify the user
func (mpc *micPayChan) microPay() {

	check := mpc.accBook.createPayment(mpc.wallet)

	if err := mpc.conn.WriteJsonMsg(check); err != nil {
		fmt.Println(err)
		return
	}

	ack := &core.PayChanAck{}
	if err := mpc.conn.ReadJsonMsg(ack); err != nil {
		fmt.Println(err)
		return
	}

	if ack.RechargeRes.NextAction == core.PayResultSuccess {
		if err := mpc.accBook.setNewReceipt(ack.RechargeRes); err != nil {
			fmt.Println(err)
			return
		}
	}
}
