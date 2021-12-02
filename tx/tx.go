package tx

type TransactionInfo struct {
	To            string
	Data          []byte
	WalletAddress string
	PrivateKeyStr string
}
