package payment

import (
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/pangolink/miner-pool/account"
	"io/ioutil"
	"sync"
)

type PacketPaymentProtocol interface {
	WalletAddr() (string, string)
	OpenPacketWallet(auth string) error
	SetupAesConn(string) (account.CryptConn, error)
}

type AccountBook struct {
	sync.RWMutex
	Counter  int
	Nonce    int
	UnSettle int64
}

type PacketAccountant struct {
	*leveldb.DB
	*AccountBook
}

func initAccountant(rPath string) (*PacketAccountant, error) {
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

	pa := &PacketAccountant{
		DB:          db,
		AccountBook: nil,
	}

	return pa, nil
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
	sWallet    *SafeWallet
	wallet     account.Wallet
	accountant *PacketAccountant
}

func InitProtocol(wPath, rPath string) (PacketPaymentProtocol, error) {

	ac, err := initAccountant(rPath)
	if err != nil {
		return nil, err
	}

	sw, err := initWallet(wPath)
	if err != nil {
		fmt.Println("[PPP] InitProtocol initWallet err:", err)
		sw = &SafeWallet{}
	}

	pw := &PacketWallet{
		sWallet:    sw,
		accountant: ac,
	}
	return pw, nil
}
