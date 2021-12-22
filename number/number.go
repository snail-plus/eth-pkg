package number

import (
	"math"
	"math/big"
)

// 计算原始ETH 余额
func FormatEthValue(in *big.Int) *big.Float {
	return FormatValue(in, 18)
}

// 计算指定精度余额
func FormatValue(in *big.Int, n int) *big.Float {
	fBalance := new(big.Float)
	fBalance.SetString(in.String())
	ethValue := new(big.Float).Quo(fBalance, big.NewFloat(math.Pow10(n)))
	return ethValue
}
