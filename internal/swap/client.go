package swap

import (
	"context"
	"crypto/ecdsa"
	"filthy/internal/account"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

type EthBasedClient struct {
	Client         *ethclient.Client
	PrivateKey     *ecdsa.PrivateKey
	PublicKeyECDSA *ecdsa.PublicKey
	Address        common.Address
	ChainID        *big.Int
	Transactor     *bind.TransactOpts
	Nonce          *big.Int
}

func NewBasedClient(client *ethclient.Client, wallet account.Wallet) EthBasedClient {
	privateKey := wallet.PrivateKey

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("error casting public key to ECDSA")
		return EthBasedClient{}
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return EthBasedClient{}
	}

	transactor, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return EthBasedClient{}
	}

	ethBasedClientTemp := EthBasedClient{
		Client:         client,
		PrivateKey:     privateKey,
		PublicKeyECDSA: publicKeyECDSA,
		Address:        address,
		ChainID:        chainID,
		Transactor:     transactor,
	}

	return ethBasedClientTemp
}

func (ethBasedClient *EthBasedClient) PendingNonceUint64() uint64 {
	// calculate next nonce
	nonce, _ := ethBasedClient.Client.PendingNonceAt(context.Background(), ethBasedClient.Address)
	//errorsutil.HandleError(nonceErr)
	return nonce
}
