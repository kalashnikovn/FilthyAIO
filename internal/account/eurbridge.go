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

type EurBridgeAccount struct {
	Wallet
	Helper
}

func NewEurBridgeAccount(wallet Wallet) EurBridgeAccount {
	return EurBridgeAccount{
		wallet,
		Helper{},
	}
}

func (acc EurBridgeAccount) Bridge() {
	constants.Logger.Warn(fmt.Sprintf("Начинаю Eur Bridge для аккаунта %s", acc.PublicKey.String()))
	utils.SendTelegramMessage(fmt.Sprintf("🟨 Начинаю Eur Bridge для аккаунта %s", acc.PublicKey.String()))

	moneyChain, err := acc.FindChainWithMoney("EUR", constants.SETTINGS.EurB.FromNetworks, acc.PublicKey)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Ошибка при поиске сети с балансом EUR")
		utils.SendTelegramMessage("🟥 Ошибка при поиске сети с балансом EUR\n\n", "Ошибка: ", err.Error())
		errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
		}
		return
	}

	balance := moneyChain.Balance

	if balance.UiAmount == 0 {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Сеть"):       moneyChain.Chain,
			color.InRed("Баланс EUR"): balance.UiAmount,
		}).Error("Недостаточный баланс EUR на аккаунте")
		utils.SendTelegramMessage("🟥 Недостаточный баланс EUR на аккаунте\n\n", "Сеть: ", moneyChain.Chain, "\nБаланс EUR: ", balance.UiAmount)
		errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
		}
		return
	}

	chainFrom := moneyChain.Chain
	randomTo, err := utils.RandomChain(constants.SETTINGS.EurB.ToNetworks, chainFrom)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Сеть, где лежат деньги"):     chainFrom,
			color.InRed("Сеть, куда нужно отправить"): chainFrom,
			color.InRed("Ошибка"):                     err,
		}).Error("Ошибка при получении рандомной сети для бриджа...")
		errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
		}
		return
	}

	scan := constants.SCANS[chainFrom]

	constants.Logger.WithFields(logrus.Fields{
		color.InPurple("Из сети"):    chainFrom,
		color.InPurple("В сеть"):     randomTo,
		color.InPurple("Баланс EUR"): balance.UiAmount,
	}).Trace("Найдена сеть с балансом EUR. Начинаю Eur.b бридж")
	utils.SendTelegramMessage("🟪 Найдена сеть с балансом EUR. Начинаю Eur.b бридж\n\n",
		"Баланс EUR: ", balance.UiAmount,
		"\nИз сети: ", chainFrom,
		"\nВ сеть: ", randomTo)

	swapArgs := acc.GetAmounts(balance, "eurb")

	//spew.Dump(swapArgs)

	amount := swapArgs.Amount

	allowance, err := acc.Allowance(chainFrom, "EUR", "eurb", acc.Wallet)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Ошибка при получении Allowance на EUR")
		utils.SendTelegramMessage("🟥 Ошибка при получении Allowance на EUR\n\n",
			"Ошибка: ", err.Error())
		errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
		}
		return
	}
	cmp := amount.Cmp(allowance)
	if cmp == 1 {
		constants.Logger.WithFields(logrus.Fields{
			color.InYellow("Amount"):    amount,
			color.InYellow("Allowance"): allowance,
		}).Warn("Требуется апрув на EUR")
		utils.SendTelegramMessage("🟪 Требуется апрув на EUR\n\n",
			"Amount: ", amount,
			"\nAllowance: ", allowance)

		approve, errApprove := acc.Approve(chainFrom, amount, "EUR", "eurb", acc.Wallet)
		if errApprove != nil {
			constants.Logger.WithFields(logrus.Fields{
				color.InRed("Ошибка"): errApprove,
			}).Error("Не удалось сделать апрув")
			utils.SendTelegramMessage("🟥 Не удалось сделать апрув\n\n",
				"Ошибка: ", errApprove.Error())
			errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
			if errWrite != nil {
				constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
			}
			return
		}
		if approve.Status == 0 {
			constants.Logger.WithFields(logrus.Fields{
				color.InRed("Хэш транзакции"): approve.TxHash,
				color.InRed("Ссылка на скан"): scan + approve.TxHash.String(),
			}).Error("Транзакция на апрув не прошла")
			utils.SendTelegramMessage("🟥 Транзакция на апрув не прошла\n\n",
				"Ссылка на скан: ", scan+approve.TxHash.String())
			errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
			if errWrite != nil {
				constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
			}
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

	swap, err := acc.EurBridgeSwap(chainFrom, randomTo, swapArgs)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
			//"Debug":  swapArgs,
		}).Error("Не удалось отправить транзакцию на бридж")
		utils.SendTelegramMessage("🟥 Не удалось отправить транзакцию на бридж\n\n",
			"Ошибка: ", err.Error())
		errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
		}
		return
	}

	if swap.Status == 0 {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Хэш транзакции"): swap.TxHash,
			color.InRed("Ссылка на скан"): scan + swap.TxHash.String(),
		}).Error("Транзакция на бридж не прошла")
		utils.SendTelegramMessage("🟥 Транзакция на бридж не прошла\n\n",
			"Ссылка на скан: ", scan+swap.TxHash.String())
		errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
		}
		return
	}

	constants.Logger.WithFields(logrus.Fields{
		color.InGreen("Хэш транзакции"): swap.TxHash,
		color.InGreen("Ссылка на скан"): scan + swap.TxHash.String(),
	}).Info("Транзакция на бридж успешно смайнилась")
	utils.SendTelegramMessage("🟩 Транзакция на бридж успешно смайнилась\n\n",
		"Ссылка на скан: ", scan+swap.TxHash.String())

	delay := acc.GetRandomActivityDelay()

	constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующей активностью", delay))
	utils.SendTelegramMessage(fmt.Sprintf("Сплю %d секунд перед следующей активностью", delay))
	time.Sleep(time.Duration(delay) * time.Second)

}

func (acc EurBridgeAccount) EurBridgeSwap(chainFrom, chainTo string, args SwapArgs) (*types.Receipt, error) {
	client := constants.CLIENTS[chainFrom]

	nonce, err := client.PendingNonceAt(context.Background(), acc.PublicKey)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["eurb"][chainFrom])
	contractAbi, err := ReadAbi(constants.EUR_BRIDGE_ABI)
	if err != nil {
		return nil, err
	}

	//from := acc.PublicKey
	dstChainId := constants.STARGATE_CHAIN_ID[chainTo]

	toAddress := acc.PublicKey.Bytes()

	//toAddress := common.LeftPadBytes(acc.PublicKey.Bytes(), 32)
	//fmt.Println("Padded address:", common.Bytes2Hex(paddedAddress))

	amount := args.Amount

	refundAddress := acc.PublicKey
	zroPaymentAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	adapterParams := common.Hex2Bytes(acc.GetAdapterParams())

	encodedData, err := contractAbi.Pack("send",
		dstChainId, toAddress, amount, refundAddress, zroPaymentAddress, adapterParams)
	if err != nil {
		return nil, err
	}

	value, err := acc.GetSwapValue(chainFrom, chainTo, amount)
	if err != nil {
		return nil, err
	}

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

	gasLimit, err := GetGasLimit(client, msg)
	//spew.Dump(gasPrice, gasLimit)
	if err != nil {
		return nil, err
	}

	canSwap := ValidateFee(chainFrom, gasPrice, gasLimit, value)

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

	//fmt.Println("Bridge ", signedTx.Hash())

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		constants.Logger.Debug(gasPrice, gasLimit)
		constants.Logger.Debug(*args.Amount, *args.MinAmount)
		return nil, err
	}

	hash, err := WaitMined(client, signedTx)
	if err != nil {
		return nil, err
	}

	return hash, nil

}

func (acc EurBridgeAccount) GetSwapValue(chainFrom, chainTo string, amount *big.Int) (*big.Int, error) {
	client := constants.CLIENTS[chainFrom]

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["eurb"][chainFrom])

	contractAbi, err := ReadAbi(constants.EUR_BRIDGE_ABI)
	if err != nil {
		return nil, err
	}

	dstChainId := constants.STARGATE_CHAIN_ID[chainTo]
	toAddress := acc.PublicKey.Bytes()

	adapterParams := common.Hex2Bytes(acc.GetAdapterParams())

	encodedData, err := contractAbi.Pack("estimateSendFee",
		dstChainId, toAddress, amount, true, adapterParams)
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

	res, err := contractAbi.Unpack("estimateSendFee", result)
	if err != nil {
		return nil, err
	}

	return res[0].(*big.Int), nil
}

func (acc EurBridgeAccount) GetAdapterParams() string {
	return "000100000000000000000000000000000000000000000000000000000000000493e0"
}
