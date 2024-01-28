package account

import (
	"context"
	"errors"
	"filthy/internal/constants"
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

type CoreBridgeAccount struct {
	Wallet
	Helper
}

func NewCoreBridgeAccount(wallet Wallet) CoreBridgeAccount {
	return CoreBridgeAccount{
		wallet,
		Helper{},
	}
}

func (acc CoreBridgeAccount) Bridge() {
	constants.Logger.Warn(fmt.Sprintf("Начинаю CoreDao Bridge для аккаунта %s", acc.PublicKey.String()))
	utils.SendTelegramMessage(fmt.Sprintf("Начинаю CoreDao Bridge для аккаунта %s", acc.PublicKey.String()))

	chain := []string{"bsc"}

	moneyChain, err := acc.FindChainWithMoney("USDC", chain, acc.PublicKey)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Ошибка при поиске сети с балансом USDT")
		utils.SendTelegramMessage("🟥 Ошибка при поиске сети с балансом USDT\n\n", "Ошибка: ", err.Error())

		return
	}
	minAmount := constants.SETTINGS.Accounts.MinUsdcBalance
	balance := moneyChain.Balance

	if balance.UiAmount < minAmount {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Сеть"):        moneyChain.Chain,
			color.InRed("Баланс USDT"): balance.UiAmount,
		}).Error("Максимальный баланс в сети меньше минимального количества")
		utils.SendTelegramMessage("🟥 Максимальный баланс в сети меньше минимального количества\n\n", "Сеть: ", moneyChain.Chain, "\nБаланс USDT: ", balance.UiAmount)

		return
	}

	chainFrom := moneyChain.Chain

	//spew.Dump(randomTo)

	constants.Logger.WithFields(logrus.Fields{
		color.InPurple("Из сети"):     chainFrom,
		color.InPurple("В сеть"):      "Core DAO",
		color.InPurple("Баланс USDT"): balance.UiAmount,
	}).Trace("Найдена сеть с балансом USDT. Начинаю CoreDao бридж")
	utils.SendTelegramMessage("🟪 Найдена сеть с балансом USDT. Начинаю CoreDao бридж\n\n",
		"Баланс USDT: ", balance.UiAmount,
		"\nИз сети: ", chainFrom,
		"\nВ сеть: ", "bsc")

	swapArgs := acc.GetAmounts(balance, "coredao")

	//spew.Dump(swapArgs)
	amount := swapArgs.Amount
	//spew.Dump(swapArgs)

	allowance, err := acc.Allowance(chainFrom, "USDC", "coredao", acc.Wallet)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Ошибка при получении Allowance на USDT")
		utils.SendTelegramMessage("🟥 Ошибка при получении Allowance на USDT\n\n",
			"Ошибка: ", err.Error())
		return
	}

	scan := constants.SCANS[chainFrom]

	cmp := amount.Cmp(allowance)
	if cmp == 1 {
		//fmt.Println("need approve")
		constants.Logger.WithFields(logrus.Fields{
			color.InYellow("Amount"):    amount,
			color.InYellow("Allowance"): allowance,
		}).Warn("Требуется апрув на USDT")
		utils.SendTelegramMessage("🟪 Требуется апрув на USDT\n\n",
			"Amount: ", amount,
			"\nAllowance: ", allowance)

		approve, errApprove := acc.Approve(chainFrom, amount, "USDC", "coredao", acc.Wallet)
		if errApprove != nil {
			constants.Logger.WithFields(logrus.Fields{
				color.InRed("Ошибка"): errApprove,
			}).Error("Не удалось сделать апрув")
			utils.SendTelegramMessage("🟥 Не удалось сделать апрув\n\n",
				"Ошибка: ", errApprove.Error())
			return
		}

		if approve.Status == 0 {
			constants.Logger.WithFields(logrus.Fields{
				color.InRed("Хэш транзакции"): approve.TxHash,
				color.InRed("Ссылка на скан"): scan + approve.TxHash.String(),
			}).Error("Транзакция на апрув не прошла")
			utils.SendTelegramMessage("🟥 Транзакция на апрув не прошла\n\n",
				"Ссылка на скан: ", scan+approve.TxHash.String())
			return
		}

		constants.Logger.WithFields(logrus.Fields{
			color.InGreen("Хэш транзакции"): approve.TxHash,
			color.InGreen("Ссылка на скан"): scan + approve.TxHash.String(),
		}).Info("Транзакция на апрув успешно смайнилась")
		utils.SendTelegramMessage("🟩 Транзакция на апрув успешно смайнилась\n\n",
			"Ссылка на скан: ", scan+approve.TxHash.String())

		delay := acc.GetRandomActivityDelay()

		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующей транзакцией", delay))
		utils.SendTelegramMessage(fmt.Sprintf("Сплю %d секунд перед следующей транзакцией", delay))

		time.Sleep(time.Duration(delay) * time.Second)
	}

	constants.Logger.Warn("Отравляю транзакцию на мост...")

	swap, err := acc.CoreBridgeSwap(swapArgs)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
			//"Debug":  swapArgs,
		}).Error("Не удалось отправить транзакцию на бридж")
		utils.SendTelegramMessage("🟥 Не удалось отправить транзакцию на бридж\n\n",
			"Ошибка: ", err.Error())
		return
	}

	if swap.Status == 0 {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Хэш транзакции"): swap.TxHash,
			color.InRed("Ссылка на скан"): scan + swap.TxHash.String(),
		}).Error("Транзакция на бридж не прошла")
		utils.SendTelegramMessage("🟥 Транзакция на бридж не прошла\n\n",
			"Ссылка на скан: ", scan+swap.TxHash.String())
		return
	}

	constants.Logger.WithFields(logrus.Fields{
		color.InGreen("Хэш транзакции"): swap.TxHash,
		color.InGreen("Ссылка на скан"): scan + swap.TxHash.String(),
	}).Info("Транзакция на бридж успешно смайнилась")
	utils.SendTelegramMessage("🟩 Транзакция на бридж успешно смайнилась\n\n",
		"Ссылка на скан: ", scan+swap.TxHash.String())

	//delay := acc.GetRandomActivityDelay()
	//
	//constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующей активностью", delay))
	//time.Sleep(time.Duration(delay) * time.Second)

}

func (acc CoreBridgeAccount) CoreBridgeSwap(args SwapArgs) (*types.Receipt, error) {
	chainFrom := "bsc"
	client := constants.CLIENTS[chainFrom]

	nonce, err := client.PendingNonceAt(context.Background(), acc.PublicKey)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["coredao"]["bsc"])
	contractAbi, err := utils.ReadAbi(constants.CORE_DAO_TO_ABI)
	if err != nil {
		return nil, err
	}

	token := common.HexToAddress(constants.CONTRACTS["USDC"]["bsc"])
	//spew.Dump(token)
	amountLD := args.Amount
	//amountLD := big.NewInt(100000000000000000)
	to := acc.PublicKey

	callParams := struct {
		RefundAddress     common.Address
		ZroPaymentAddress common.Address
	}{
		RefundAddress:     acc.PublicKey,
		ZroPaymentAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
	}
	adapterParams := []byte{}

	encodedData, err := contractAbi.Pack("bridge",
		token, amountLD, to, callParams, adapterParams)
	if err != nil {
		return nil, err
	}

	value, err := acc.GetSwapValue()
	if err != nil {
		return nil, err
	}

	//spew.Dump(tokenId, value, sum)

	msg := ethereum.CallMsg{
		To:    &contractAddress,
		From:  acc.PublicKey,
		Value: value,
		Data:  encodedData,
	}

	//gasPrice, err := client.SuggestGasPrice(context.Background())
	gasPrice, err := GetGasPrice(client, chainFrom)
	if err != nil {
		return nil, err
	}

	gasLimit, err := utils.GetGasLimit(client, msg)
	//spew.Dump(gasPrice, gasLimit)
	if err != nil {
		return nil, err
	}

	canSwap := utils.ValidateFee(chainFrom, gasPrice, gasLimit, value)

	if canSwap == false {
		return nil, errors.New("полная комиссия моста выше указанного лимита")
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
		return nil, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		constants.Logger.Debug(gasPrice, gasLimit)
		return nil, err
	}

	hash, err := utils.WaitMined(client, signedTx)
	if err != nil {
		return nil, err
	}

	return hash, nil

}

func (acc CoreBridgeAccount) GetSwapValue() (*big.Int, error) {
	client := constants.CLIENTS["bsc"]

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["coredao"]["bsc"])

	contractAbi, err := utils.ReadAbi(constants.CORE_DAO_TO_ABI)
	if err != nil {
		return nil, err
	}

	adapterParams := []byte{}

	encodedData, err := contractAbi.Pack("estimateBridgeFee",
		true, adapterParams)
	if err != nil {
		return nil, err
	}

	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress,
		Data: encodedData,
	}, nil)
	if err != nil {
		return nil, err
	}

	res, err := contractAbi.Unpack("estimateBridgeFee", result)
	if err != nil {
		return nil, err
	}

	return res[0].(*big.Int), nil
}
