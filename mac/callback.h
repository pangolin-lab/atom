#ifndef _CALL_BACK_HEADER_
#define _CALL_BACK_HEADER_

        typedef void (*BlockChainDataSyncNotifier) (int, char*);
        void bridge_data_func(BlockChainDataSyncNotifier f , int t, char* v);

        typedef void (*SystemActionCallBack) (int, char*);
        void bridge_sys_func(SystemActionCallBack f, int t, char* v);

        enum CallBackActionType{
            BalanceSynced,
            TxProcessTips,
            SubPoolSynced,
            MarketPoolSynced,
        };

#endif