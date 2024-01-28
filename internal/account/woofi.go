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
	"math/rand"
	"time"
)

type WoofiBridgeAccount struct {
	Wallet
	Helper
}

func NewWoofiBridgeAccount(wallet Wallet) WoofiBridgeAccount {
	return WoofiBridgeAccount{
		wallet,
		Helper{},
	}
}

func (acc WoofiBridgeAccount) Bridge() {
	constants.Logger.Warn(fmt.Sprintf("Начинаю WooFi Bridge для аккаунта %s", acc.PublicKey.String()))
	utils.SendTelegramMessage(fmt.Sprintf("🟨 Начинаю WooFi Bridge для аккаунта %s", acc.PublicKey.String()))

	moneyChain, err := acc.FindChainWithMoney("USDC", constants.SETTINGS.WooFi.FromNetworks, acc.PublicKey)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Ошибка при поиске сети с балансом USDC")
		utils.SendTelegramMessage("🟥 Ошибка при поиске сети с балансом USDC\n\n", "Ошибка: ", err.Error())
		errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
		}
		return
	}

	minAmount := constants.SETTINGS.Accounts.MinUsdcBalance
	balance := moneyChain.Balance

	if balance.UiAmount < minAmount {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Сеть"):        moneyChain.Chain,
			color.InRed("Баланс USDC"): balance.UiAmount,
		}).Error("Максимальный баланс в сети меньше минимального количества")
		utils.SendTelegramMessage("🟥 Максимальный баланс в сети меньше минимального количества\n\n", "Сеть: ", moneyChain.Chain, "\nБаланс USDC: ", balance.UiAmount)
		errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
		}
		return
	}

	chainFrom := moneyChain.Chain
	randomTo, err := utils.RandomChain(constants.SETTINGS.WooFi.ToNetworks, chainFrom)
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

	constants.Logger.WithFields(logrus.Fields{
		color.InPurple("Из сети"):     chainFrom,
		color.InPurple("В сеть"):      randomTo,
		color.InPurple("Баланс USDC"): balance.UiAmount,
	}).Trace("Найдена сеть с балансом USDC. Начинаю WooFi бридж")
	utils.SendTelegramMessage("🟪 Найдена сеть с балансом USDC. Начинаю WooFi бридж\n\n",
		"Баланс USDC: ", balance.UiAmount,
		"\nИз сети: ", chainFrom,
		"\nВ сеть: ", randomTo)

	swapArgs := acc.GetAmounts(balance, "woofi")

	amount := swapArgs.Amount

	allowance, err := acc.Allowance(chainFrom, "USDC", "woofi", acc.Wallet)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Ошибка при получении Allowance на USDC")
		utils.SendTelegramMessage("🟥 Ошибка при получении Allowance на USDC\n\n",
			"Ошибка: ", err.Error())
		errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
		}
		return
	}
	scan := constants.SCANS[chainFrom]
	cmp := amount.Cmp(allowance)
	if cmp == 1 {

		constants.Logger.WithFields(logrus.Fields{
			color.InYellow("Amount"):    amount,
			color.InYellow("Allowance"): allowance,
		}).Warn("Требуется апрув на USDC")
		utils.SendTelegramMessage("🟪 Требуется апрув на USDC\n\n",
			"Amount: ", amount,
			"\nAllowance: ", allowance)

		approve, errApprove := acc.Approve(chainFrom, amount, "USDC", "woofi", acc.Wallet)
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

	swap, err := acc.WoofiBridgeSwap(chainFrom, randomTo, swapArgs)
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

func (acc WoofiBridgeAccount) WoofiBridgeSwap(chainFrom, chainTo string, args SwapArgs) (*types.Receipt, error) {
	client := constants.CLIENTS[chainFrom]

	nonce, err := client.PendingNonceAt(context.Background(), acc.PublicKey)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["woofi"][chainFrom])
	contractAbi, err := ReadAbi(constants.WOOFI_ABI)
	if err != nil {
		return nil, err
	}

	refId := acc.GetRefID()

	to := acc.PublicKey

	fromToken := common.HexToAddress(constants.CONTRACTS["USDC"][chainFrom])
	toToken := common.HexToAddress(constants.CONTRACTS["USDC"][chainTo])

	chainId := constants.STARGATE_CHAIN_ID[chainTo]

	srcInfos := struct {
		FromToken       common.Address
		BridgeToken     common.Address
		FromAmount      *big.Int
		MinBridgeAmount *big.Int
	}{
		FromToken:       fromToken,
		BridgeToken:     fromToken,
		FromAmount:      args.Amount,
		MinBridgeAmount: args.Amount,
	}

	dstInfos := struct {
		ChainId             uint16
		ToToken             common.Address
		BridgeToken         common.Address
		MinToAmount         *big.Int
		AirdropNativeAmount *big.Int
	}{
		ChainId:             chainId,
		ToToken:             toToken,
		BridgeToken:         toToken,
		MinToAmount:         args.MinAmount,
		AirdropNativeAmount: big.NewInt(0),
	}

	encodedData, err := contractAbi.Pack("crossSwap",
		refId, to, srcInfos, dstInfos)
	if err != nil {
		return nil, err
	}

	value, err := acc.GetSwapValue(chainFrom, chainTo, args)
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

func (acc WoofiBridgeAccount) GetSwapValue(chainFrom, chainTo string, args SwapArgs) (*big.Int, error) {
	client := constants.CLIENTS[chainFrom]

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["woofi"][chainFrom])

	contractAbi, err := ReadAbi(constants.WOOFI_ABI)
	if err != nil {
		return nil, err
	}

	refId := acc.GetRefID()

	to := acc.PublicKey

	fromToken := common.HexToAddress(constants.CONTRACTS["USDC"][chainFrom])
	toToken := common.HexToAddress(constants.CONTRACTS["USDC"][chainTo])

	chainId := constants.STARGATE_CHAIN_ID[chainTo]

	srcInfos := struct {
		FromToken       common.Address
		BridgeToken     common.Address
		FromAmount      *big.Int
		MinBridgeAmount *big.Int
	}{
		FromToken:       fromToken,
		BridgeToken:     fromToken,
		FromAmount:      args.Amount,
		MinBridgeAmount: args.Amount,
	}

	dstInfos := struct {
		ChainId             uint16
		ToToken             common.Address
		BridgeToken         common.Address
		MinToAmount         *big.Int
		AirdropNativeAmount *big.Int
	}{
		ChainId:             chainId,
		ToToken:             toToken,
		BridgeToken:         toToken,
		MinToAmount:         args.MinAmount,
		AirdropNativeAmount: big.NewInt(0),
	}

	encodedData, err := contractAbi.Pack("quoteLayerZeroFee",
		refId, to, srcInfos, dstInfos)
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

	res, err := contractAbi.Unpack("quoteLayerZeroFee", result)
	if err != nil {
		return nil, err
	}

	return res[0].(*big.Int), nil
}

func (acc WoofiBridgeAccount) GetRefID() *big.Int {
	rand.Seed(time.Now().UnixNano())
	result := int(float64(100_000)*rand.Float64()) + int(time.Now().UnixNano()/1e6)
	return big.NewInt(int64(result))
}
