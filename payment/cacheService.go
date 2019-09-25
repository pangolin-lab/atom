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
	PoolAddressInMarket = []byte("_DB_Pool_Address_In_Market_")
	PoolDetailsCached   = []byte("_DB_Pool_Details_Information_Cached_")
)

const (
	DBKeySubPoolArr = "_DB_SUB_POOL_ARR_"

	_ = iota
)

type DataSyncCallBack interface {
	SubPoolDataSynced()
	MarketPoolDataSynced()
	DataSyncedFailed(err error)
}

type BlockChainDataService struct {
	sync.RWMutex
	*leveldb.DB
	callBack DataSyncCallBack

	PoolsInMarket    []common.Address
	MySubscribedPool []common.Address

	poolDetails    map[string]*ethereum.PoolDetail
	channelDetails map[string]*ethereum.PayChannel
}

func InitBlockDataCache(dataPath, mainAddr string, cb DataSyncCallBack) (*BlockChainDataService, error) {
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

	cachedDetail := make(map[string]*ethereum.PoolDetail)
	_ = utils.GetObj(db, PoolDetailsCached, cachedDetail)

	addresses := make([]common.Address, 0)
	_ = utils.GetObj(db, PoolAddressInMarket, addresses)

	bcd := &BlockChainDataService{
		DB:               db,
		callBack:         cb,
		poolDetails:      cachedDetail,
		PoolsInMarket:    addresses,
		MySubscribedPool: make([]common.Address, 0),
		channelDetails:   make(map[string]*ethereum.PayChannel),
	}
	go bcd.SyncPacketMarket()

	if mainAddr != "" {
		key := fmt.Sprintf("%s%s", DBKeySubPoolArr, mainAddr)
		pools := make([]common.Address, 0)
		if err := utils.GetObj(bcd.DB, []byte(key), pools); err != nil {
			fmt.Println("[InitBlockDataCache]  load cached subscribed pool warning:", err)
		}
		bcd.MySubscribedPool = pools
		go bcd.SyncSubscribedPool(mainAddr)
	}
	fmt.Println("[InitBlockDataCache] init success......")
	return bcd, nil
}

func (bcd *BlockChainDataService) SyncPacketMarket() {
	addresses := ethereum.PoolAddressList()
	if addresses == nil {
		fmt.Println("[DataService] ethereum PoolAddressList no data found:")
		return
	}

	if err := utils.SaveObj(bcd.DB, PoolAddressInMarket, addresses); err != nil {
		fmt.Println("[DataService] ethereum save pool address in market err:", err)
		return
	}
	bcd.PoolsInMarket = addresses
	fmt.Println("[dataService] SyncPacketMarket success......")
	bcd.callBack.MarketPoolDataSynced()
}

func (bcd *BlockChainDataService) LoadDetailsOfArr(addrArr []common.Address) []*ethereum.PoolDetail {
	poolArr := make([]*ethereum.PoolDetail, 0)
	for _, addr := range addrArr {
		p, e := bcd.LoadPoolDetails(addr.String())
		if e != nil {
			fmt.Println(e)
			continue
		}

		poolArr = append(poolArr, p)
	}
	return poolArr
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

func (bcd *BlockChainDataService) SyncSubscribedPool(addr string) {
	if addr == "" {
		return
	}
	poolArr, err := ethereum.MySubPools(addr)
	if err != nil {
		fmt.Println("[DataService] ethereum MySubPools err:", err)
		return
	}

	bcd.Lock()
	defer bcd.Unlock()
	bcd.MySubscribedPool = poolArr
	key := fmt.Sprintf("%s%s", DBKeySubPoolArr, addr)
	if err := utils.SaveObj(bcd.DB, []byte(key), poolArr); err != nil {
		fmt.Println("[DataService] wallet info save err:", err)
	}

	bcd.callBack.SubPoolDataSynced()
}

func (bcd *BlockChainDataService) Finalized() {
	bcd.Close()
}
