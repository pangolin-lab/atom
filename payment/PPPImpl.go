package payment

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
	if pw.wallet == nil || !pw.wallet.IsOpen() {
		return false
	}

	if pw.Chanel == nil || pw.pool.MainAddr != poolAddr {
		return false
	}

	return true
}

func (pw *PacketWallet) OpenPayChannel(errCh chan error, pool *ethereum.PoolDetail, auth string) error {
	if pw.wallet == nil || !pw.wallet.IsOpen() {
		if err := pw.openWallet(auth); err != nil {
			return err
		}
	}

	if pw.Chanel != nil {
		pw.CloseChannel()
	}

	conn, err := pw.connectToMiner(pool.Seeds)
	if err != nil {
		return err
	}

	bootInfo, err := pw.handshake(conn)
	if err != nil {
		return err
	}

	miner, err := pw.randomMiner(bootInfo.MinerIDs)
	if err != nil {
		return err
	}

	accBook, err := pw.checkAccount(pool.MainAddr, bootInfo.Sig, bootInfo.LatestReceipt)
	if err != nil {
		return err
	}

	pw.Chanel = &Chanel{
		pool:    pool,
		miner:   miner,
		conn:    conn,
		accBook: accBook,
	}

	go pw.monitor(errCh)

	return nil
}

func (pw *PacketWallet) SetupAesConn(target string) (account.CryptConn, error) {

	miner := pw.miner

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
		UserSubAddr: pw.sWallet.SubAddr,
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
	return account.NewAesConn(conn, pw, aesKey[:], iv)
}

func (pw *PacketWallet) WalletAddr() (string, string) {
	return pw.sWallet.MainAddr, pw.sWallet.SubAddr
}
