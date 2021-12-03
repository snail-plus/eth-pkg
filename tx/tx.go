package tx

import "math/big"

type TransactionInfo struct {
	To            string
	Data          []byte
	WalletAddress string
	PrivateKeyStr string
	Value         *big.Int
}
