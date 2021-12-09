package tx

import (
	"github.com/ethereum/go-ethereum/params"
	"math/big"
)

type GasProvider interface {
	GetGasPrice(contractFunc string) *big.Int

	GetGasLimit(contractFunc string) *big.Int
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

func GetGasProvider() *DefaultGasProvider {
	return &DefaultGasProvider{
		gasPrice: big.NewInt(params.GWei * 5),
		gasLimit: big.NewInt(3000000),
	}
}
