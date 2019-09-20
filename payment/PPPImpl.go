package payment

import "C"
import (
	"crypto/aes"
	"crypto/rand"
	"fmt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/utils"
	acc "github.com/pangolink/go-node/account"
	"github.com/pangolink/go-node/network"
	"github.com/pangolink/go-node/service/rpcMsg"
	"github.com/pangolink/miner-pool/account"
	"io"
)

func (pw *PacketWallet) IsPayChannelOpen(poolAddr string) bool {
	return pw.isChanOpen() && pw.isWalletOpen() && pw.payChan.pool.MainAddr == poolAddr
}

func (pw *PacketWallet) OpenPayChannel(errCh chan error, pool *ethereum.PoolDetail, auth string) error {
	pw.errCh = errCh

	if pw.wallet == nil || !pw.wallet.IsOpen() {
		if err := pw.openWallet(auth); err != nil {
			return err
		}
	}

	if pw.isChanOpen() {
		pw.CloseChannel()
	}

	c, err := pw.createChan(pool)
	if err != nil {
		return err
	}
	pw.payChan = c
	go pw.monitor()
	return nil
}

func (pw *PacketWallet) SetupAesConn(target string) (account.CryptConn, error) {

	if !pw.isChanOpen() {
		if err := pw.tryReopen(); err != nil {
			return nil, err
		}
	}

	miner := pw.payChan.miner

	conn, err := utils.GetSavedConn(miner.NetAddr)
	if err != nil {
		fmt.Printf("\nConnect to miner failed:[%s]", err.Error())
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	io.ReadFull(rand.Reader, iv[:])

	jsonConn := &network.JsonConn{Conn: conn}
	req := rpcMsg.AesConnSetup{
		IV:          iv[:],
		Target:      target,
		UserSubAddr: pw.accBook.SubAddr,
	}

	req.Sig = pw.wallet.SignSub(req)
	if err := jsonConn.Syn(req); err != nil {
		fmt.Println("Send salt to miner failed:", err)
		return nil, err
	}

	aesKey := new(acc.PipeCryptKey)
	if err := acc.GenerateAesKey(aesKey, miner.ID.ToPubKey(), pw.wallet.CryptKey()); err != nil {
		return nil, fmt.Errorf("[SetupAesConn] error aeskey")
	}
	return account.NewAesConn(conn, pw.accBook, aesKey[:], iv)
}

func (pw *PacketWallet) Finish() {
	//TODO::
}

func (pw *PacketWallet) SyncWalletData() *Accountant {
	go pw.accBook.synBalance(pw.database, pw.callBack)
	return pw.accBook
}

func (pw *PacketWallet) Wallet(auth string) (account.Wallet, error) {
	if pw.wallet != nil && pw.wallet.IsOpen() {
		return pw.wallet, nil
	}
	if err := pw.openWallet(auth); err != nil {
		return nil, err
	}
	return pw.wallet, nil
}
