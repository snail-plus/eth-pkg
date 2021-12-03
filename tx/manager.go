package tx

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/snail-plus/eth-pkg/client"
	"math/big"
	"sync"
)

type TransactionManager interface {
	// 返回hash
	ExecuteTransaction(to string, data string, value *big.Int, gasPrice *big.Int, gasLimit *big.Int) (string, error)
	GetNonce(ctx context.Context, account common.Address)
}

type FastRawTransactionManager struct {
	nonce         uint64
	mutex         sync.Mutex
	web3Client    *client.Web3Client
	walletAddress string
	privateKeyStr string
}

func (f FastRawTransactionManager) ExecuteTransaction(to string, data string, value *big.Int,
	gasPrice *big.Int, gasLimit uint64) (string, error) {
	ctx := context.Background()
	nonce := f.GetNonce(ctx)

	txInfo := TransactionInfo{
		To:            to,
		Data:          []byte(data),
		WalletAddress: f.walletAddress,
		PrivateKeyStr: f.privateKeyStr,
		Value:         value,
	}

	signTx, err := f.web3Client.SignNewTxInfo(txInfo, nonce, gasPrice, gasLimit)
	if err != nil {
		return "", err
	}

	txHash, err := f.web3Client.SendTransaction(ctx, signTx)
	return txHash, err

}

func (f FastRawTransactionManager) GetNonce(ctx context.Context) uint64 {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.nonce == 0 {
		nonce, err := f.web3Client.GetNonce(ctx, f.walletAddress)
		if err != nil {
			return 1
		}

		f.nonce = nonce
	} else {
		f.nonce = f.nonce + 1
	}

	return f.nonce
}
