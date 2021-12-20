package secure

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func StringToPrivateKey(privateKeyStr string) (*ecdsa.PrivateKey, error) {
	privateKeyByte, err := hexutil.Decode(privateKeyStr)
	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.ToECDSA(privateKeyByte)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func PrivateKeyToAddressStr(privateKeyStr string) (string, error) {
	address, err := PrivateKeyToAddress(privateKeyStr)
	if err != nil {
		return "", err
	}

	return address.String(), nil
}

func PrivateKeyToAddress(privateKeyStr string) (common.Address, error) {
	key, err := StringToPrivateKey(privateKeyStr)
	if err != nil {
		return common.Address{}, err
	}

	address := crypto.PubkeyToAddress(key.PublicKey)
	return address, nil
}
