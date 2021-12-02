package client

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/snail-plus/eth-pkg/secure"
	"github.com/snail-plus/eth-pkg/tx"
	"math/big"
	"sync"
)

var once sync.Once

var rpcClient *rpc.Client
var err error

var ethClient *ethclient.Client

func initClient(nodeUrl string) {
	rpcClient, err = rpc.Dial(nodeUrl)
	if err != nil {
		panic(err)
	}

	ethClient = ethclient.NewClient(rpcClient)
	fmt.Println("初始化client成功")
}

func SendTransaction(ctx context.Context, tx *types.Transaction) (string, error) {
	data, err := tx.MarshalBinary()
	if err != nil {
		return "", err
	}

	var result string
	err = rpcClient.CallContext(ctx, &result, "eth_sendRawTransaction", hexutil.Encode(data))
	return result, err
}

func GetEthClient(nodeUrl string) *ethclient.Client {
	once.Do(func() {
		initClient(nodeUrl)
	})
	return ethClient
}

func GetGasPrice(ctx context.Context) (*big.Int, error) {
	return ethClient.SuggestGasPrice(ctx)
}

func GetNonce(ctx context.Context, walletAddress string) (uint64, error) {
	return ethClient.NonceAt(ctx, common.HexToAddress(walletAddress), nil)
}

func NetworkID() (*big.Int, error) {
	return ethClient.NetworkID(context.Background())
}

func GetSigner() types.Signer {
	chainID, _ := NetworkID()
	signer := types.LatestSignerForChainID(chainID)
	return signer
}

func TransactionByHash(ctx context.Context, hashStr string) (tx *types.Transaction, isPending bool, err error) {
	hash := common.HexToHash(hashStr)
	return ethClient.TransactionByHash(ctx, hash)
}

func SignNewTx(ctx context.Context, txInfo tx.TransactionInfo) (*types.Transaction, error) {
	key, err := secure.StringToPrivateKey(txInfo.PrivateKeyStr)
	if err != nil {
		return nil, err
	}

	nonce, err := GetNonce(ctx, txInfo.WalletAddress)
	if err != nil {
		return nil, err
	}

	gasPrice, err := GetGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	toAddress := common.HexToAddress(txInfo.To)

	tx, err := types.SignNewTx(key, GetSigner(), &types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    big.NewInt(0),
		Gas:      600000,
		GasPrice: gasPrice,
		Data:     txInfo.Data,
	})

	return tx, err
}
