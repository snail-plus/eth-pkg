package tx

import (
	"context"
	"github.com/snail-plus/eth-pkg/secure"
	"log"
	"math/big"
	"sync"
	"time"
)

type TransactionManager interface {
	// 返回hash
	ExecuteTransaction(to string, data []byte, value *big.Int, gasPrice *big.Int, gasLimit uint64) (string, error)
	GetNonce(ctx context.Context, account string, refresh bool) (uint64, error)
	GetPrivateKey() string
}

type FastRawTransactionManager struct {
	nonce               uint64
	refreshNonceTime    int64
	lastAccessNonceTime int64
	mutex               *sync.RWMutex
	web3Client          *Web3Client
	privateKeyStr       string
}

func NewDefaultTransactionManager(web3Client *Web3Client,
	privateKeyStr string) TransactionManager {
	txManager := &FastRawTransactionManager{
		nonce:         0,
		mutex:         new(sync.RWMutex),
		web3Client:    web3Client,
		privateKeyStr: privateKeyStr,
	}

	timer := time.NewTicker(10 * time.Second)
	go func() {
		for range timer.C {

			func() {
				txManager.mutex.Lock()
				defer txManager.mutex.Unlock()

				txManager.syncNonce()
			}()

		}
	}()

	txManager.syncNonce()
	return txManager
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
	return txHash, err

}

func (f *FastRawTransactionManager) GetNonce(ctx context.Context, account string, refresh bool) (uint64, error) {
	f.mutex.Lock()

	defer func() {
		f.lastAccessNonceTime = time.Now().Unix()
		f.mutex.Unlock()
	}()

	// 注意 交易在交易池 中 节点nonce 不会增加(执行后记录当前nonce) 这里需要客户端自己计算 然后定时更新nonce
	if time.Now().Unix()-f.refreshNonceTime >= 60 {
		refresh = true
	}

	if f.nonce == 0 || refresh || f.refreshNonceTime == 0 {
		f.syncNonce()
	} else if time.Now().Unix()-f.refreshNonceTime < 10 {
		// 定时任务在没有 访问nonce 30 秒以外 会定时更新nonce 这里直接返回即可
		f.nonce = f.nonce + 1
		return f.nonce, nil
	} else {
		return f.nonce, nil
	}

	return f.nonce, nil
}

// 外部调用加锁
func (f *FastRawTransactionManager) syncNonce() error {
	addressStr, _ := secure.PrivateKeyToAddressStr(f.privateKeyStr)
	nonce, err := f.web3Client.GetNonce(context.Background(), addressStr)
	if err != nil {
		return err
	}
	log.Printf("syncNonce, address: %s, nonce: %d", addressStr, nonce)
	f.nonce = nonce
	f.refreshNonceTime = time.Now().Unix()
	return nil
}

func (f *FastRawTransactionManager) GetPrivateKey() string {
	return f.privateKeyStr
}
