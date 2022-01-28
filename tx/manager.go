package tx

import (
	"context"
	"github.com/snail-plus/eth-pkg/secure"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"
)

type TransactionManager interface {
	// 返回hash
	ExecuteTransaction(to string, data []byte, value *big.Int, gasPrice *big.Int, gasLimit uint64) (string, error)
	GetNonce(ctx context.Context, account string, refresh bool) (uint64, error)
}

type FastRawTransactionManager struct {
	nonce            uint64
	refreshNonceTime int64
	mutex            sync.Mutex
	web3Client       *Web3Client
	privateKeyStr    string
}

func NewDefaultTransactionManager(web3Client *Web3Client,
	privateKeyStr string) TransactionManager {
	return &FastRawTransactionManager{
		nonce:         0,
		mutex:         sync.Mutex{},
		web3Client:    web3Client,
		privateKeyStr: privateKeyStr,
	}
}

func (f *FastRawTransactionManager) ExecuteTransaction(to string, data []byte, value *big.Int,
	gasPrice *big.Int, gasLimit uint64) (string, error) {
	ctx := context.Background()
	addressStr, err := secure.PrivateKeyToAddressStr(f.privateKeyStr)
	if err != nil {
		return "", err
	}

	nonce, err := f.GetNonce(ctx, addressStr, false)
	if err != nil {
		return "", err
	}

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
	if err != nil && strings.Contains(err.Error(), "nonce too") {
		// nonce too low refresh
		f.GetNonce(ctx, addressStr, true)
	}

	return txHash, err

}

func (f *FastRawTransactionManager) GetNonce(ctx context.Context, account string, refresh bool) (uint64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if time.Now().Unix()-f.refreshNonceTime >= 60 {
		refresh = true
	}

	if f.nonce == 0 || refresh {
		nonce, err := f.web3Client.GetNonce(ctx, account)
		if err != nil {
			return 0, err
		}
		log.Printf("address: %s, nonce: %d", account, nonce)
		f.nonce = nonce
		f.refreshNonceTime = time.Now().Unix()
	} else {
		f.nonce = f.nonce + 1
	}

	return f.nonce, nil
}
