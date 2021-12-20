package tx

import (
	"context"
	"math/big"
	"sync"
)

type TransactionManager interface {
	// 返回hash
	ExecuteTransaction(to string, data []byte, value *big.Int, gasPrice *big.Int, gasLimit uint64) (string, error)
	GetNonce(ctx context.Context, account string) uint64
}

type FastRawTransactionManager struct {
	nonce         uint64
	mutex         sync.Mutex
	web3Client    *Web3Client
	walletAddress string
	privateKeyStr string
}

func NewDefaultTransactionManager(web3Client *Web3Client,
	walletAddress string, privateKeyStr string) TransactionManager {
	return &FastRawTransactionManager{
		nonce:         0,
		mutex:         sync.Mutex{},
		web3Client:    web3Client,
		walletAddress: walletAddress,
		privateKeyStr: privateKeyStr,
	}
}

func (f *FastRawTransactionManager) ExecuteTransaction(to string, data []byte, value *big.Int,
	gasPrice *big.Int, gasLimit uint64) (string, error) {
	ctx := context.Background()
	nonce := f.GetNonce(ctx, f.walletAddress)

	txInfo := TransactionInfo{
		To:            to,
		Data:          data,
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

func (f *FastRawTransactionManager) GetNonce(ctx context.Context, account string) uint64 {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.nonce == 0 {
		nonce, err := f.web3Client.GetNonce(ctx, account)
		if err != nil {
			return 1
		}

		f.nonce = nonce
	} else {
		f.nonce = f.nonce + 1
	}

	return f.nonce
}
