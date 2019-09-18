package payment

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolink/miner-pool/account"
	"github.com/pangolink/miner-pool/core"
	"sync"
)

type BookItem struct {
	sync.RWMutex
	Counter  int
	UnSettle int64
	Balance  int64
}

type AccBook struct {
	*BookItem
	receipt string
	payer   string
	key     string
	db      *leveldb.DB
}

func loadAccBook(path, mainAddr, poolAddr string) (*AccBook, error) {
	opts := opt.Options{
		ErrorIfExist: true,
		Strict:       opt.DefaultStrict,
		Compression:  opt.NoCompression,
		Filter:       filter.NewBloomFilter(10),
	}
	db, err := leveldb.OpenFile(path, &opts)
	if err != nil {
		return nil, err
	}

	ab := &AccBook{
		db:       db,
		receipt:  poolAddr,
		payer:    mainAddr,
		key:      fmt.Sprintf("%s%s%s", mainAddr, AccBookKeyJoin, poolAddr),
		BookItem: &BookItem{},
	}

	data, err := db.Get([]byte(ab.key), nil)
	if err != nil {
		if err != leveldb.ErrNotFound {
			return nil, err
		}
		return ab, nil
	}

	if err := json.Unmarshal(data, ab.BookItem); err != nil {
		return nil, err
	}

	return ab, nil
}

func (ac *AccBook) incrUsage(n int) {
	ac.Lock()
	defer ac.Unlock()
	ac.Counter += n
}

func (ac *AccBook) createPayment(w account.Wallet) *core.PayChanReq {
	ac.RLock()
	defer ac.RUnlock()

	recharge := &core.MicroPay{
		Recipient: ac.receipt,
		Usage:     ac.Counter,
		Contract:  ethereum.Conf.MicroPaySys,
		UnSettled: ac.UnSettle,
		Balance:   ac.Balance,
	}

	sig, _ := w.Sign(recharge)
	req := &core.PayChanReq{
		MsgType:  core.Recharge,
		Sig:      sig,
		Recharge: recharge,
	}
	return req
}

func (ac *AccBook) setNewReceipt(receipt *core.MicroReceipt) error {
	ac.Lock()
	defer ac.Unlock()

	n := receipt.Recharged
	ac.Counter -= n
	ac.UnSettle += int64(n)
	data, err := json.Marshal(ac.BookItem)
	if err != nil {
		return err
	}

	if err := ac.db.Put([]byte(ac.key), data, nil); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
