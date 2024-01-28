package swap

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	erc20 "github.com/jerrychan807/1inch-trading-bot/contracts/ERC20"
	go1inch "github.com/jon4hz/go-1inch"
)

func GetTokenInstance(contractAddr common.Address, client *ethclient.Client) *erc20.Erc20 {
	instance, err := erc20.NewErc20(contractAddr, client)
	if err != nil {
		//log.Fatal(err)
		return nil
	}

	return instance
}

func GetTokenDecimals(instance *erc20.Erc20) uint8 {
	decimals, err := instance.Decimals(&bind.CallOpts{})
	if err != nil {
		//log.Fatal(err)
		return 0
	}
	return decimals
}

func GetApproveAllowance(tokenAddr string, walletAddr string, network string) (string, error) {
	client := go1inch.NewClient()
	res, _, err := client.ApproveAllowance(context.Background(), go1inch.Network(network), tokenAddr, walletAddr)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//Logger.WithFields(logrus.Fields{"Allowance": res.Allowance}).Info("Allowance for 1inchRouter")
	return res.Allowance, err
}

func GetApproveTx(tokenAddr string, network string) (*go1inch.ApproveTransactionRes, error) {
	client := go1inch.NewClient()
	res, _, err := client.ApproveTransaction(
		context.Background(),
		go1inch.Network(network),
		tokenAddr,
		nil,
	)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//Logger.WithFields(logrus.Fields{"ToAddr": res.To, "Data": res.Data}).Info("")
	//spew.Dump(res)
	return res, nil
}

func GetSwapTxData(tokenAddr string, toTokenAddr string, amount string, walletAddr string, slippage int64, network string) (*go1inch.SwapRes, error) {
	client := go1inch.NewClient()

	res, _, err := client.Swap(context.Background(), go1inch.Network(network),
		tokenAddr,
		toTokenAddr,
		amount,
		walletAddr,
		slippage,
		nil,
	)
	if err != nil {
		//fmt.Println(err)
		return nil, err
	}
	return res, nil
}
