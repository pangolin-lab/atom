package payment

import (
	"fmt"
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pangolin-lab/atom/ethereum"
	"github.com/pangolin-lab/atom/utils"
	"sync"
)

var (
	MarketDataVersion   = []byte("_DB_MARKET_DATA_VERSION_")
	PoolAddressInMarket = []byte("_DB_Pool_Address_In_Market_")
	PoolDetailsCached   = []byte("_DB_Pool_Details_Information_Cached_")
)

const (
	DBKeySubPoolArr = "_DB_SUB_POOL_Arr_"
	DBMicPayChan    = "_DB_MICRO_PAY_CHANNEL_"
)

type BlockChainDataService struct {
	sync.RWMutex
	*leveldb.DB

	marketDataVersion uint32
	poolsOfMyChannel  []common.Address
	poolsInMarket     []common.Address
	poolDetails       map[string]*ethereum.PoolDetail
	channelDetails    map[string]*ethereum.PayChannel
}

func InitBlockDataCache(dataPath, mainAddr string) (*BlockChainDataService, error) {
	opts := opt.Options{
		Strict:      opt.DefaultStrict,
		Compression: opt.NoCompression,
		Filter:      filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(dataPath, &opts)
	if err != nil {
		return nil, err
	}
	fmt.Println("[InitBlockDataCache] open block chain cached database success......")

	dataVersion := uint32(0)
	if data, err := db.Get(MarketDataVersion, nil); err == nil {
		dataVersion = utils.ByteToUint(data)
	}

	cachedDetail := make(map[string]*ethereum.PoolDetail)
	_ = utils.GetObj(db, PoolDetailsCached, cachedDetail)

	bcd := &BlockChainDataService{
		DB:                db,
		marketDataVersion: dataVersion,
		poolDetails:       cachedDetail,
		poolsOfMyChannel:  make([]common.Address, 0),
		poolsInMarket:     make([]common.Address, 0),
		channelDetails:    make(map[string]*ethereum.PayChannel),
	}

	go bcd.loadPacketMarket()

	if mainAddr != "" {
		go bcd.loadSubPools(mainAddr)
	}
	fmt.Println("[InitBlockDataCache] init success......")
	return bcd, nil
}

func (bcd *BlockChainDataService) loadPacketMarket() {
	addresses := make([]common.Address, 0)
	if err := utils.GetObj(bcd.DB, PoolAddressInMarket, addresses); err != nil {
		bcd.syncPacketMarket()
		return
	}
	bcd.poolsInMarket = addresses
	fmt.Println("[dataService] loadPacketMarket success......")
	go bcd.syncPacketMarket()
}

func (bcd *BlockChainDataService) syncPacketMarket() {
	newVer := ethereum.MarketDataVersion()
	if bcd.marketDataVersion == newVer {
		fmt.Println("[DataService-syncPacketMarket]  no need to sync packet market data")
		return
	}

	addresses := ethereum.PoolAddressList()
	if addresses == nil {
		fmt.Println("[DataService] ethereum PoolAddressList no data found:")
		return
	}

	if err := utils.SaveObj(bcd.DB, PoolAddressInMarket, addresses); err != nil {
		fmt.Println("[DataService] ethereum save pool address in market err:", err)
		return
	}
	bcd.poolsInMarket = addresses
	bcd.marketDataVersion = newVer
	_ = bcd.Put(MarketDataVersion, utils.UintToByte(newVer), nil)
	fmt.Println("[dataService] syncPacketMarket success......")
}

func (bcd *BlockChainDataService) loadSubPools(addr string) {
	key := fmt.Sprintf("%s%s", DBKeySubPoolArr, addr)
	pools := make([]common.Address, 0)
	if err := utils.GetObj(bcd.DB, []byte(key), pools); err != nil || len(pools) == 0 {
		bcd.SyncSubPools(addr)
		return
	}

	bcd.Lock()
	defer bcd.Unlock()
	bcd.poolsOfMyChannel = pools
}

func (bcd *BlockChainDataService) SyncSubPools(addr string) {
	poolArr, err := ethereum.MySubPools(addr)
	if err != nil {
		fmt.Println("[DataService] ethereum MySubPools err:", err)
		return
	}

	key := fmt.Sprintf("%s%s", DBKeySubPoolArr, addr)
	if err := utils.SaveObj(bcd.DB, []byte(key), poolArr); err != nil {
		fmt.Println("[DataService] SyncSubPools err:", err)
	}

	bcd.Lock()
	defer bcd.Unlock()
	bcd.poolsOfMyChannel = poolArr
}

func (bcd *BlockChainDataService) LoadPoolDetails(poolAddr string) (*ethereum.PoolDetail, error) {
	bcd.RLock()
	if details, ok := bcd.poolDetails[poolAddr]; ok {
		bcd.RUnlock()
		return details, nil
	}
	bcd.RUnlock()

	d, e := ethereum.PoolDetails(common.HexToAddress(poolAddr))
	if e != nil {
		return nil, e
	}

	go bcd.synDetailsCache(poolAddr, d)
	return d, nil
}

func (bcd *BlockChainDataService) synDetailsCache(addr string, d *ethereum.PoolDetail) {
	bcd.Lock()
	defer bcd.Unlock()
	bcd.poolDetails[addr] = d

	if err := utils.SaveObj(bcd.DB, PoolDetailsCached, bcd.poolDetails); err != nil {
		fmt.Println("[DataService]  synDetailsCache err:", err)
	}
}

func (bcd *BlockChainDataService) PoolsInMarket() []common.Address {
	bcd.RLock()
	defer bcd.RUnlock()

	bcd.syncPacketMarket()

	return bcd.poolsInMarket
}
