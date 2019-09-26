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
	PoolDetailsCached    = []byte("_DB_Pool_Details_Information_Cached_")
	ChannelDetailsCached = []byte("_DB_Channel_Details_Information_Cached_")
)

type DataSyncCallBack interface {
	ChannelInfoSynced()
	MarketPoolDataSynced()
	DataSyncedFailed(err error)
}

type BlockChainDataService struct {
	sync.RWMutex

	*leveldb.DB
	callBack DataSyncCallBack

	poolVer        uint32
	channelVer     uint32
	PoolDetails    map[string]*ethereum.PoolDetail
	ChannelDetails map[string]*ethereum.ChannelDetail
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

	pools := make(map[string]*ethereum.PoolDetail)
	_ = utils.GetObj(db, PoolDetailsCached, pools)

	channels := make(map[string]*ethereum.ChannelDetail)
	_ = utils.GetObj(db, ChannelDetailsCached, channels)

	bcd := &BlockChainDataService{
		DB:             db,
		callBack:       cb,
		PoolDetails:    pools,
		ChannelDetails: channels,
	}

	go bcd.SyncPacketMarket()

	if mainAddr != "" {
		go bcd.SyncMyChannelDetails(mainAddr)
	}
	fmt.Println("[InitBlockDataCache] init success......")
	return bcd, nil
}

func (bcd *BlockChainDataService) SyncPacketMarket() {
	newVer := ethereum.MarketDataVersion()
	if bcd.poolVer == newVer {
		fmt.Println("[DataService] ethereum no market data changed:")
		return
	}

	addresses := ethereum.PoolAddressList()
	if err := utils.SaveObj(bcd.DB, PoolDetailsCached, addresses); err != nil {
		fmt.Println("[DataService] ethereum save pool address in market err:", err)
		return
	}

	bcd.Lock()
	defer bcd.Unlock()
	bcd.PoolDetails = make(map[string]*ethereum.PoolDetail)

	for _, addr := range addresses {
		p, e := ethereum.PoolDetails(addr)
		if e != nil {
			fmt.Println(e)
			continue
		}
		bcd.PoolDetails[addr.String()] = p
	}

	_ = utils.SaveObj(bcd.DB, PoolDetailsCached, bcd.PoolDetails)
	fmt.Println("[dataService] SyncPacketMarket success......")
	bcd.callBack.MarketPoolDataSynced()
}

func (bcd *BlockChainDataService) SyncMyChannelDetails(addr string) {
	newVer := ethereum.MyChannelVersion(addr)
	if bcd.channelVer == newVer && newVer != 0 {
		return
	}

	poolArr, err := ethereum.MySubPools(addr)
	if err != nil {
		fmt.Println("[DataService] ethereum MySubPools err:", err)
		return
	}

	bcd.Lock()
	defer bcd.Unlock()
	bcd.ChannelDetails = make(map[string]*ethereum.ChannelDetail)

	for _, pool := range poolArr {
		d, err := ethereum.GetChanDetails(common.HexToAddress(addr), pool)
		if err != nil {
			fmt.Println("[DataService] Channel details err:", err)
			continue
		}
		bcd.ChannelDetails[pool.String()] = d
	}

	_ = utils.SaveObj(bcd.DB, ChannelDetailsCached, bcd.ChannelDetails)
	fmt.Println("[dataService] SyncMyChannelDetails success......")
	bcd.callBack.ChannelInfoSynced()
}

func (bcd *BlockChainDataService) Finalized() {
	bcd.Close()
}
