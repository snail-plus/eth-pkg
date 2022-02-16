package tx

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/snail-plus/eth-pkg/secure"
	"github.com/snail-plus/goutil/maputil"
	"log"
	"math/big"
	"strings"
)

type Web3Client struct {
	ethClient  *ethclient.Client
	rpcClient  *rpc.Client
	chainId    *big.Int
	gethClient *gethclient.Client
}

func NewWeb3Client(nodeUrl string) *Web3Client {

	rpcClient, err := rpc.Dial(nodeUrl)
	if err != nil {
		panic(err)
	}

	ethClient := ethclient.NewClient(rpcClient)
	networkID, err := ethClient.NetworkID(context.Background())
	if err != nil {
		panic(err)
	}

	return &Web3Client{
		ethClient:  ethClient,
		rpcClient:  rpcClient,
		chainId:    networkID,
		gethClient: gethclient.New(rpcClient),
	}
}

func (e *Web3Client) GetEthClient() *ethclient.Client {
	return e.ethClient
}

func (e *Web3Client) GetGEthClient() *gethclient.Client {
	return e.gethClient
}

func (e *Web3Client) SendTransaction(ctx context.Context, tx *types.Transaction) (string, error) {
	data, err := tx.MarshalBinary()
	if err != nil {
		return "", err
	}

	var result string
	err = e.rpcClient.CallContext(ctx, &result, "eth_sendRawTransaction", hexutil.Encode(data))
	return result, err
}

func (e *Web3Client) GetGasPrice(ctx context.Context) (*big.Int, error) {
	return e.ethClient.SuggestGasPrice(ctx)
}

func (e *Web3Client) GetNonce(ctx context.Context, walletAddress string) (uint64, error) {
	return e.ethClient.NonceAt(ctx, common.HexToAddress(walletAddress), big.NewInt(-1))
}

func (e *Web3Client) NetworkID() (*big.Int, error) {
	return e.ethClient.NetworkID(context.Background())
}

func (e *Web3Client) GetSigner() types.Signer {
	signer := types.LatestSignerForChainID(e.chainId)
	return signer
}

func (e *Web3Client) TransactionByHash(ctx context.Context, hashStr string) (tx *types.Transaction, isPending bool, err error) {
	hash := common.HexToHash(hashStr)
	return e.ethClient.TransactionByHash(ctx, hash)
}

func (e *Web3Client) TraceTransaction(ctx context.Context, hashStr string) (result interface{}, err error) {
	err = e.rpcClient.CallContext(ctx, &result, "debug_traceTransaction", common.HexToHash(hashStr))
	return result, err
}

func (e *Web3Client) SignNewTx(ctx context.Context, txInfo TransactionInfo) (*types.Transaction, error) {
	key, err := secure.StringToPrivateKey(txInfo.PrivateKeyStr)
	if err != nil {
		return nil, err
	}

	walletAddress := crypto.PubkeyToAddress(key.PublicKey)
	nonce, err := e.GetNonce(ctx, walletAddress.String())
	if err != nil {
		return nil, err
	}

	gasPrice, err := e.GetGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	var toAddress *common.Address
	if txInfo.To != "" {
		tmpToAddress := common.HexToAddress(txInfo.To)
		toAddress = &tmpToAddress
	}

	tx, err := types.SignNewTx(key, e.GetSigner(), &types.LegacyTx{
		Nonce:    nonce,
		To:       toAddress,
		Value:    big.NewInt(0),
		Gas:      600000,
		GasPrice: gasPrice,
		Data:     txInfo.Data,
	})

	return tx, err
}

func (e *Web3Client) SignNewTxInfo(txInfo TransactionInfo,
	nonce uint64, gasPrice *big.Int, gas uint64) (*types.Transaction, error) {
	key, err := secure.StringToPrivateKey(txInfo.PrivateKeyStr)
	if err != nil {
		return nil, err
	}

	var toAddress *common.Address
	if txInfo.To != "" {
		tmpToAddress := common.HexToAddress(txInfo.To)
		toAddress = &tmpToAddress
	}

	tx, err := types.SignNewTx(key, e.GetSigner(), &types.LegacyTx{
		Nonce:    nonce,
		To:       toAddress,
		Value:    txInfo.Value,
		Gas:      gas,
		GasPrice: gasPrice,
		Data:     txInfo.Data,
	})

	return tx, err
}

func (e *Web3Client) NewPendingTransactionFilter() (string, error) {
	pendingTransactionFilter := NewPendingTransactionFilter(e.rpcClient)
	filterID, err := pendingTransactionFilter.GetFilterId()
	return filterID, err
}

func (e *Web3Client) NewLogFilter(filterQuery FilterQuery) (string, error) {
	filter := NewLogFilterFilter(e.rpcClient, filterQuery)
	filterID, err := filter.GetFilterId()
	return filterID, err
}

// pending Filter 时返回 数组hash 获取LOG 时候返回Log对象  0 or nil means latest block -1 pending
func (e *Web3Client) EthLogFlowable(filterQuery FilterQuery, pullInterval int64) chan interface{} {

	filter := NewLogFilterFilter(e.rpcClient, filterQuery)
	filter.Run(pullInterval)

	logChan := filter.LogChan
	return logChan
}

// pending Filter 时返回 数组hash 获取LOG 时候返回Log对象  0 or nil means latest block -1 pending
func (e *Web3Client) EthPendingFlowable(pullInterval int64) chan interface{} {

	filter := NewPendingTransactionFilter(e.rpcClient)
	filter.Run(pullInterval)

	logChan := filter.LogChan
	return logChan
}

func (e *Web3Client) ParityAllTransactions(ctx context.Context) ([]*RPCTransaction, error) {
	var result []*RPCTransaction
	err := e.rpcClient.CallContext(ctx, &result, "parity_allTransactions")
	return result, err
}

func (e *Web3Client) SubscribePendingTransactions(ctx context.Context, ch chan *types.Transaction) (*rpc.ClientSubscription, error) {
	hashChan := make(chan common.Hash, cap(ch))
	subscription, err := e.gethClient.SubscribePendingTransactions(ctx, hashChan)

	go func() {
		// 控制并发查询
		var sem = make(chan int, 3)

		for {
			select {
			case <-ctx.Done():
				break
			case err = <-subscription.Err():
				break
			case txHah := <-hashChan:

				sem <- 1

				go func() {
					defer func() {
						<-sem
					}()

					pendingTx, _, err := e.ethClient.TransactionByHash(ctx, txHah)
					if err != nil {
						log.Printf("TransactionByHash error: %s", err.Error())
						return
					}

					ch <- pendingTx
				}()

			}
		}
	}()

	return subscription, err

}

// content := map[string]map[string]map[string]*RPCTransaction{
//		"pending": make(map[string]map[string]*RPCTransaction),
//		"queued":  make(map[string]map[string]*RPCTransaction),
//	}
func (e *Web3Client) TxPoolContent(ctx context.Context) (map[string]map[string]map[string]*RPCTransaction, error) {
	var result map[string]map[string]map[string]*RPCTransaction
	err := e.rpcClient.CallContext(ctx, &result, "txpool_content")
	if err != nil {
		return result, err
	}

	return result, nil
}

func (e *Web3Client) TxPoolContentPending(ctx context.Context, filter func(toAddress string) bool) ([]*RPCTransaction, error) {
	var result map[string]map[string]map[string]*RPCTransaction
	err := e.rpcClient.CallContext(ctx, &result, "txpool_content")
	if err != nil {
		return nil, err
	}

	pending := result["pending"]
	queued := result["queued"]

	var fullTxArr []*RPCTransaction

	flatTx := func(txMap map[string]map[string]*RPCTransaction) []*RPCTransaction {
		values := maputil.Values(txMap)
		var pendingTxArr []*RPCTransaction

		for _, v := range values {
			txMap := v.(map[string]*RPCTransaction)
			txArr := maputil.Values(txMap)
			for _, txItem := range txArr {
				pendingTx := txItem.(*RPCTransaction)

				if filter == nil {
					pendingTxArr = append(pendingTxArr, pendingTx)
					continue
				}

				if filter(strings.ToLower(pendingTx.To.String())) {
					pendingTxArr = append(pendingTxArr, pendingTx)
				}

			}
		}

		return pendingTxArr
	}

	fullTxArr = append(fullTxArr, flatTx(pending)...)
	fullTxArr = append(fullTxArr, flatTx(queued)...)

	return fullTxArr, nil
}
