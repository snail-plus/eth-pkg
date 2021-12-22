package tx

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"log"
	"math/big"
	"sync/atomic"
	"time"
)

type GasProvider interface {
	GetGasPrice(contractFunc string) *big.Int

	GetGasLimit(contractFunc string) *big.Int
}

type DynamicGasProvider struct {
	gasPrice  atomic.Value
	gasLimit  *big.Int
	ethClient *ethclient.Client
	ticker    *time.Ticker
}

func (d *DynamicGasProvider) GetGasPrice(contractFunc string) *big.Int {
	return d.gasPrice.Load().(*big.Int)
}

func (d *DynamicGasProvider) GetGasLimit(contractFunc string) *big.Int {
	return d.gasLimit
}

type DefaultGasProvider struct {
	gasPrice *big.Int
	gasLimit *big.Int
}

func (d *DefaultGasProvider) GetGasPrice(contractFunc string) *big.Int {
	return d.gasPrice
}

func (d *DefaultGasProvider) GetGasLimit(contractFunc string) *big.Int {
	return d.gasLimit
}

func NewGasProvider() GasProvider {
	return &DefaultGasProvider{
		gasPrice: big.NewInt(params.GWei * 5),
		gasLimit: big.NewInt(3000000),
	}
}

func NewDynamicGasProvider(ethClient *ethclient.Client) GasProvider {
	var gasValue atomic.Value
	gasValue.Store(big.NewInt(params.GWei * 5))

	provider := &DynamicGasProvider{
		gasPrice:  gasValue,
		gasLimit:  big.NewInt(3000000),
		ethClient: ethClient,
		ticker:    time.NewTicker(1 * time.Minute),
	}

	go func() {
		for range provider.ticker.C {
			price, err := provider.ethClient.SuggestGasPrice(context.Background())
			if err != nil {
				log.Printf("get gasPrice error: %s", err.Error())
				continue
			}

			provider.gasPrice.Store(price)
		}
	}()
	return provider
}
