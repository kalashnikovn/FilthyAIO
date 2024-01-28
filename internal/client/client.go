package client

import (
	"context"
	"fmt"
	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/ethereum/go-ethereum/ethclient"
)

func New(rpc string) *ethclient.Client {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		fmt.Println("Проблема с рпс:", rpc)
		panic(err)
	}

	return client
}

func NewAptos(rpc string) *aptosclient.RestClient {
	client, err := aptosclient.Dial(context.Background(), rpc)
	if err != nil {
		fmt.Println("Проблема с рпс:", rpc)
		panic(err)
	}

	return client
}
