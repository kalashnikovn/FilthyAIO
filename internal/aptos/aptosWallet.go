package aptos

import (
	"bufio"
	"encoding/hex"
	"filthy/internal/account"
	"github.com/coming-chat/go-aptos/aptosaccount"
	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/ethereum/go-ethereum/crypto"
	"os"
	"strings"
)

type AptosPair struct {
	Wallet      account.Wallet
	Aptos       *aptosaccount.Account
	AptosClient *aptosclient.RestClient
}

type Keys struct {
	Evm   string
	Aptos string
}

func New(pks Keys, client *aptosclient.RestClient) (AptosPair, error) {
	evm, err := NewEVM(pks.Evm)
	if err != nil {
		return AptosPair{}, err
	}
	aptos, err := NewAptos(pks.Aptos)
	if err != nil {
		return AptosPair{}, err
	}

	return AptosPair{
		Wallet:      evm,
		Aptos:       aptos,
		AptosClient: client,
	}, nil
}

func NewEVM(pk string) (account.Wallet, error) {
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return NewEVM(pk[2:])
	}

	publicKey := crypto.PubkeyToAddress(privateKey.PublicKey)

	return account.Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func NewAptos(pk string) (*aptosaccount.Account, error) {
	privateKey, err := hex.DecodeString(pk)
	if err != nil {
		return NewAptos(pk[2:])
	}
	account := aptosaccount.NewAccount(privateKey)

	return account, nil
}

func NewWallets(client *aptosclient.RestClient) ([]AptosPair, error) {
	keys, err := readAptosPairKeys("aptosKeys.txt")
	if err != nil {
		return nil, err
	}

	accs, err := getWallets(keys, client)
	if err != nil {
		return nil, err
	}

	return accs, nil
}

func getWallets(pks []Keys, client *aptosclient.RestClient) ([]AptosPair, error) {
	var accounts []AptosPair
	for _, pk := range pks {
		account, err := New(pk, client)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)

	}
	return accounts, nil
}

func readAptosPairKeys(filePath string) ([]Keys, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	keys := []Keys{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		key := Keys{Evm: strings.TrimSpace(parts[0]), Aptos: strings.TrimSpace(parts[1])}
		keys = append(keys, key)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}
