package refuel

import (
	"context"
	"filthy/internal/account"
	"filthy/internal/constants"
	"filthy/internal/swap"
	"filthy/internal/utils"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"math/big"
	"time"
)

type RefuelAccount struct {
	account.Wallet
	account.Helper
}

func NewRefuelAccount(wallet account.Wallet) RefuelAccount {
	return RefuelAccount{
		wallet,
		account.Helper{},
	}
}

func (acc RefuelAccount) Refuel(chainsTo []string) {
	constants.Logger.Warn(fmt.Sprintf("Начинаю Refuel для аккаунта %s", acc.PublicKey.String()))
	chainFrom := constants.SETTINGS.Refuel.FromNetwork
	scan := constants.SCANS[chainFrom]

	for _, chain := range chainsTo {
		deposit, err := acc.Deposit(chain)
		if err != nil {
			constants.Logger.WithFields(logrus.Fields{
				color.InRed("В сеть"): chain,
				color.InRed("Ошибка"): err,
				//"Debug":  swapArgs,
			}).Error("Не удалось отправить транзакцию на рефуел")
			utils.SendTelegramMessage("🟥 Не удалось отправить транзакцию на рефуел\n\n",
				"Ошибка: ", err.Error(),
				"\nВ сеть: ", chain)

			errWrite := utils.AppendToFile("errorRefuelWallets.txt", utils.PrivateKeyToString(acc.PrivateKey))
			if errWrite != nil {
				constants.Logger.Error("Ошибка при записи приватника в файл errorRefuelWallets.txt")
			}
			return
		}

		if deposit.Status == 0 {
			constants.Logger.WithFields(logrus.Fields{
				color.InRed("Хэш транзакции"): deposit.TxHash,
				color.InRed("Ссылка на скан"): scan + deposit.TxHash.String(),
			}).Error("Транзакция на рефуел не прошла")
			utils.SendTelegramMessage("🟥 Транзакция на рефуел не прошла\n\n",
				"Ссылка на скан: ", scan+deposit.TxHash.String())

			errWrite := utils.AppendToFile("errorRefuelWallets.txt", utils.PrivateKeyToString(acc.PrivateKey))
			if errWrite != nil {
				constants.Logger.Error("Ошибка при записи приватника в файл errorRefuelWallets.txt")
			}
			return
		}

		constants.Logger.WithFields(logrus.Fields{
			color.InGreen("Хэш транзакции"): deposit.TxHash,
			color.InGreen("Ссылка на скан"): scan + deposit.TxHash.String(),
		}).Info("Транзакция на рефуел успешно смайнилась")
		utils.SendTelegramMessage("🟩 Транзакция на рефуел успешно смайнилась\n\n",
			"Ссылка на скан: ", scan+deposit.TxHash.String())

		time.Sleep(5 * time.Second)

	}

}

func (acc RefuelAccount) Deposit(chainTo string) (*types.Receipt, error) {
	chainFrom := constants.SETTINGS.Refuel.FromNetwork
	client := constants.CLIENTS[chainFrom]

	minAmount := constants.SETTINGS.Refuel.NativeMinAmount
	maxAmount := constants.SETTINGS.Refuel.NativeMaxAmount
	amount := utils.GetRandomFloat(minAmount, maxAmount)

	nonce, err := client.PendingNonceAt(context.Background(), acc.PublicKey)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(constants.REFUEL_CONTRACTS[chainFrom])
	contractAbi, err := account.ReadAbi(constants.REFUEL_ABI)
	if err != nil {
		return nil, err
	}

	destinationChainId := big.NewInt(constants.CHAIN_IDS[chainTo])
	to := acc.PublicKey

	encodedData, err := contractAbi.Pack("depositNativeToken",
		destinationChainId, to)
	if err != nil {
		return nil, err
	}

	value := swap.ToWei(amount, 18)

	msg := ethereum.CallMsg{
		To:    &contractAddress,
		From:  acc.PublicKey,
		Value: value,
		Data:  encodedData,
	}

	//gasPrice, err := client.SuggestGasPrice(context.Background())
	gasPrice, err := utils.GetGasPrice(client, chainFrom)
	if err != nil {
		return nil, err
	}

	gasLimit, err := account.GetGasLimit(client, msg)
	//spew.Dump(gasPrice, gasLimit)
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(
		nonce,
		contractAddress,
		value,
		gasLimit,
		gasPrice,
		encodedData,
	)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), acc.PrivateKey)
	if err != nil {
		//log.Fatalf("Failed to sign transaction: %v", err)
		return nil, err
	}

	//fmt.Println("Bridge ", signedTx.Hash())

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		constants.Logger.Debug(gasPrice, gasLimit)
		constants.Logger.Debug(value)
		return nil, err
	}

	hash, err := account.WaitMined(client, signedTx)
	if err != nil {
		return nil, err
	}

	return hash, nil
}
