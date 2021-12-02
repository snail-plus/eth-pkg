package client

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/snail-plus/eth-pkg/secure"
	"github.com/snail-plus/eth-pkg/tx"
	"math/big"
)

type Web3Client struct {
	ethClient *ethclient.Client
	rpcClient *rpc.Client
}

func NewWeb3Client(nodeUrl string) *Web3Client {

	rpcClient, err := rpc.Dial(nodeUrl)
	if err != nil {
		panic(err)
	}

	return &Web3Client{
		ethClient: ethclient.NewClient(rpcClient),
		rpcClient: rpcClient,
	}
}

func (e *Web3Client) GetEthClient() *ethclient.Client {
	return e.ethClient
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
	return e.ethClient.NonceAt(ctx, common.HexToAddress(walletAddress), nil)
}

func (e *Web3Client) NetworkID() (*big.Int, error) {
	return e.ethClient.NetworkID(context.Background())
}

func (e *Web3Client) GetSigner() types.Signer {
	chainID, _ := e.NetworkID()
	signer := types.LatestSignerForChainID(chainID)
	return signer
}

func (e *Web3Client) TransactionByHash(ctx context.Context, hashStr string) (tx *types.Transaction, isPending bool, err error) {
	hash := common.HexToHash(hashStr)
	return e.ethClient.TransactionByHash(ctx, hash)
}

func (e *Web3Client) SignNewTx(ctx context.Context, txInfo tx.TransactionInfo) (*types.Transaction, error) {
	key, err := secure.StringToPrivateKey(txInfo.PrivateKeyStr)
	if err != nil {
		return nil, err
	}

	nonce, err := e.GetNonce(ctx, txInfo.WalletAddress)
	if err != nil {
		return nil, err
	}

	gasPrice, err := e.GetGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	toAddress := common.HexToAddress(txInfo.To)

	tx, err := types.SignNewTx(key, e.GetSigner(), &types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    big.NewInt(0),
		Gas:      600000,
		GasPrice: gasPrice,
		Data:     txInfo.Data,
	})

	return tx, err
}
