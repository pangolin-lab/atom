package microPay

import "C"
import (
	"fmt"
	"github.com/pangolin-lab/atom/utils"
	"github.com/pangolink/go-node/network"
	"github.com/pangolink/miner-pool/account"
	"github.com/pangolink/miner-pool/core"
)

type PayChannel interface {
	Setup() error
	MinerNetAddr() string
}

type micPayChan struct {
	wallet       account.Wallet
	conn         *network.JsonConn
	minerID      *utils.PeerID
	minerNetAddr string
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

	return &micPayChan{
		wallet: w,
		conn:   &network.JsonConn{Conn: conn},
	}, nil
}

func (mpc *micPayChan) Setup() error {
	main, sub := mpc.wallet.Address()

	req := &core.ChanCreateReq{
		MainAddr: main,
		SubAddr:  sub,
	}
	sig, err := mpc.wallet.Sign(req)
	if err != nil {
		return err
	}

	syn := core.PayChanSyn{
		MsgType:   core.CreateReq,
		Sig:       sig,
		CreateReq: req,
	}

	if err := mpc.conn.WriteJsonMsg(syn); err != nil {
		return err
	}

	ack := &core.PayChanAck{}
	if err := mpc.conn.ReadJsonMsg(ack); err != nil {
		return err
	}

	if ack.Success != true {
		return fmt.Errorf("create payment channel err:%s", ack.ErrMsg)
	}

	mid, e := utils.ConvertPID(ack.MinerId)
	if e != nil {
		return utils.ErrInvalidID
	}
	mpc.minerID = mid
	mpc.minerNetAddr = mid.NetAddr()
	return nil
}

func (mpc *micPayChan) MinerNetAddr() string {
	return mpc.minerNetAddr
}
