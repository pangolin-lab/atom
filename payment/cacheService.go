package payment

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/utils"
	"sync"
)

const (
	DBKeySubPoolArr     = "_DB_SUB_POOL_Arr_"
	DBMicPayChan        = "_DB_MICRO_PAY_CHANNEL_"
	MarketDataVersion   = "_DB_MARKET_DATA_VERSION_"
	PoolAddressInMarket = "_DB_Pool_Address_In_Market_"
)

type BlockChainDataCache struct {
	sync.RWMutex
	*leveldb.DB

	marketDataVersion uint32
	poolsOfMyChannel  []common.Address
	poolsInMarket     []common.Address
	poolDetails       map[string]*ethereum.PoolDetail
	channelDetails    map[string]*ethereum.PayChannel
}

func InitBlockDataCache(dataPath, mainAddr string) (*BlockChainDataCache, error) {
	opts := opt.Options{
		ErrorIfExist: true,
		Strict:       opt.DefaultStrict,
		Compression:  opt.NoCompression,
		Filter:       filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(dataPath, &opts)
	if err != nil {
		return nil, err
	}

	dataVersion := uint32(0)
	if data, err := db.Get([]byte(MarketDataVersion), nil); err == nil {
		dataVersion = utils.ByteToUint(data)
	}

	bcd := &BlockChainDataCache{
		DB:                db,
		marketDataVersion: dataVersion,
		poolsOfMyChannel:  make([]common.Address, 0),
		poolsInMarket:     make([]common.Address, 0),
		poolDetails:       make(map[string]*ethereum.PoolDetail),
		channelDetails:    make(map[string]*ethereum.PayChannel),
	}

	go bcd.loadPacketMarket()

	if mainAddr != "" {
		go bcd.loadSubPools(mainAddr)
	}
	return bcd, nil
}

func (bcd *BlockChainDataCache) loadPacketMarket() {
	addresses := make([]common.Address, 0)
	if err := bcd.getObj([]byte(PoolAddressInMarket), addresses); err != nil {
		bcd.syncPacketMarket()
		return
	}
	bcd.poolsInMarket = addresses

	go bcd.syncPacketMarket()
}

func (bcd *BlockChainDataCache) syncPacketMarket() {
	newVer := ethereum.MarketDataVersion()
	if bcd.marketDataVersion == newVer {
		return
	}

	addresses := ethereum.PoolAddressList()
	if addresses == nil {
		fmt.Println("[PPP] ethereum PoolAddressList no data found:")
		return
	}

	if err := bcd.saveObj([]byte(PoolAddressInMarket), addresses); err != nil {
		fmt.Println("[PPP] ethereum save pool address in market err:", err)
		return
	}

	bcd.marketDataVersion = newVer
	_ = bcd.Put([]byte(MarketDataVersion), utils.UintToByte(newVer), nil)
}

func (bcd *BlockChainDataCache) loadSubPools(addr string) {
	key := fmt.Sprintf("%s%s", DBKeySubPoolArr, addr)
	pools := make([]common.Address, 0)
	if err := bcd.getObj([]byte(key), pools); err != nil || len(pools) == 0 {
		bcd.SyncSubPools(addr)
		return
	}

	bcd.Lock()
	defer bcd.Unlock()
	bcd.poolsOfMyChannel = pools
}

func (bcd *BlockChainDataCache) SyncSubPools(addr string) {
	poolArr, err := ethereum.MySubPools(addr)
	if err != nil {
		fmt.Println("[PPP] ethereum MySubPools err:", err)
		return
	}

	key := fmt.Sprintf("%s%s", DBKeySubPoolArr, addr)
	if err := bcd.saveObj([]byte(key), poolArr); err != nil {
		fmt.Println("[PPP] saveObj err:", err)
	}

	bcd.Lock()
	defer bcd.Unlock()
	bcd.poolsOfMyChannel = poolArr
}

func (bcd *BlockChainDataCache) saveObj(key []byte, v interface{}) error {

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	wo := &opt.WriteOptions{
		Sync: true,
	}

	return bcd.Put(key, data, wo)
}

func (bcd *BlockChainDataCache) getObj(key []byte, v interface{}) error {

	data, err := bcd.Get(key, nil)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, v); err != nil {
		return err
	}

	return nil
}
