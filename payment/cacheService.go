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
	DBKeySubPoolArr   = "SUB_POOL_Arr_"
	DBMicPayChan      = "MICRO_PAY_CHANNEL_"
	MarketDataVersion = "_DB_MARKET_DATA_VERSION"
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

	bcd := &BlockChainDataCache{
		DB:               db,
		poolsOfMyChannel: make([]common.Address, 0),
		poolsInMarket:    make([]common.Address, 0),
		poolDetails:      make(map[string]*ethereum.PoolDetail),
		channelDetails:   make(map[string]*ethereum.PayChannel),
	}

	go bcd.loadPacketMarket()

	if mainAddr != "" {
		go bcd.loadSubPools(mainAddr)
	}
	return bcd, nil
}

func (bcd *BlockChainDataCache) loadPacketMarket() {
	data, err := bcd.Get([]byte(MarketDataVersion), nil)
	if err != nil {
		bcd.marketDataVersion = 0
	}
	bcd.marketDataVersion = utils.ByteToUint(data)

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
		fmt.Println("[PPP] ethereum.MySubPools err:", err)
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
