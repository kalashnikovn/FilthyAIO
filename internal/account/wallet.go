package account

import (
	"bufio"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"os"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  common.Address
}

func New(pk string) (Wallet, error) {
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return New(pk[2:])
	}

	publicKey := crypto.PubkeyToAddress(privateKey.PublicKey)

	return Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func NewWallets() ([]Wallet, error) {
	keys, err := readPrivateKeys("privateKeys.txt")
	if err != nil {
		return nil, err
	}

	accs, err := getWallets(keys)
	if err != nil {
		return nil, err
	}

	return accs, nil
}

func getWallets(pks []string) ([]Wallet, error) {
	var accounts []Wallet
	for _, pk := range pks {
		account, err := New(pk)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)

	}
	return accounts, nil
}

func readPrivateKeys(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
