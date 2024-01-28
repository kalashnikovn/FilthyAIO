package swap

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"strconv"
)

func SwapTokenByInch(ethBasedClient EthBasedClient, tokenAddr string, toTokenAddr string, amount string, walletAddr string, slippage int64, network string) (*types.Receipt, error) {
	//Logger.WithFields(logrus.Fields{"walletAddr": ethBasedClient.Address}).Info("")
	//spew.Dump(ethBasedClient.Address)
	nonce := ethBasedClient.PendingNonceUint64()

	swapRes, err := GetSwapTxData(tokenAddr, toTokenAddr, amount, walletAddr, slippage, network) // 获取swap交易参数
	if err != nil {
		return nil, err
	}

	ToAddr := common.HexToAddress(swapRes.Tx.To)
	gasLimit := uint64(swapRes.Tx.Gas)

	//gasPriceInt, _ := strconv.Atoi(swapRes.Tx.GasPrice)
	//gasPrice := big.NewInt(int64(gasPriceInt))

	gasPrice, err := ethBasedClient.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	//spew.Dump(gasLimit, gasPrice)

	ethValueInt, _ := strconv.Atoi(swapRes.Tx.Value)
	ethValue := big.NewInt(int64(ethValueInt))
	data := common.FromHex(swapRes.Tx.Data)
	chainID, _ := ethBasedClient.Client.NetworkID(context.Background())

	t := types.NewTransaction(nonce, ToAddr, ethValue, gasLimit, gasPrice, data)
	s := types.NewEIP155Signer(chainID)
	signedTx, err := types.SignTx(t, s, ethBasedClient.PrivateKey)
	if err != nil {
		return nil, err
	}

	err = ethBasedClient.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		//Logger.WithFields(logrus.Fields{"err": err, "chainID": chainID}).Info("SendTransaction")
		return nil, err
	}
	//txHash := signedTx.Hash().Hex()

	tx, waitErr := bind.WaitMined(context.Background(), ethBasedClient.Client, signedTx)
	if waitErr != nil {
		//Logger.WithFields(logrus.Fields{"swapTxHash": txHash}).Info("swapTxHash Faild")
		return nil, waitErr
	}
	return tx, nil
}

func ApproveTokenByInch(ethBasedClient EthBasedClient, tokenAddr string, network string) error {
	//Logger.WithFields(logrus.Fields{"walletAddr": ethBasedClient.Address}).Info("")
	nonce := ethBasedClient.PendingNonceUint64()
	approveRes, err := GetApproveTx(tokenAddr, network)
	if err != nil {
		return err
	}

	//spew.Dump(approveRes)

	gasLimit := uint64(210000)
	ToAddr := common.HexToAddress(approveRes.To)
	ethValueInt, _ := strconv.Atoi(approveRes.Value)
	ethValue := big.NewInt(int64(ethValueInt))
	data := common.FromHex(approveRes.Data)

	//gasPriceInt, _ := strconv.Atoi(approveRes.GasPrice)
	//gasPrice := big.NewInt(int64(gasPriceInt))
	gasPrice, err := ethBasedClient.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	chainID, _ := ethBasedClient.Client.NetworkID(context.Background())
	t := types.NewTransaction(nonce, ToAddr, ethValue, gasLimit, gasPrice, data)
	s := types.NewEIP155Signer(chainID)

	signedTx, err := types.SignTx(t, s, ethBasedClient.PrivateKey)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//Logger.Info(gasPriceErr)
	err = ethBasedClient.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		//Logger.WithFields(logrus.Fields{"err": err, "chainID": chainID}).Info("SendTransaction,ApproveTokenByInch Faild")
		return err
	}
	//txHash := signedTx.Hash().Hex()

	_, waitErr := bind.WaitMined(context.Background(), ethBasedClient.Client, signedTx)
	if waitErr != nil {
		//Logger.WithFields(logrus.Fields{"approveTxHash": txHash}).Info("approveTxHash Faild")
		return err
	} else {
		//Logger.WithFields(logrus.Fields{"approveTxHash": txHash}).Info("approveTxHash Successfully")
		//spew.Dump(txHash)
	}
	return nil
}
