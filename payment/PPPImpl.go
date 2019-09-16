package payment

import (
	"fmt"
	"github.com/pangolink/miner-pool/account"
)

func (pw *PacketWallet) WalletAddr() (string, string) {
	return pw.sWallet.MainAddr, pw.sWallet.SubAddr
}

func (pw *PacketWallet) OpenPacketWallet(auth string) error {
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

func (pw *PacketWallet) SetupAesConn(target string) (account.CryptConn, error) {
	return nil, nil
}
