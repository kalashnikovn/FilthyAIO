package aptos

import (
	"context"
	"encoding/hex"
	"errors"
	"filthy/internal/account"
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

type BridgeToAptos struct {
	AptosPair
	account.Helper
}

func NewBridgeToAptos(aptosPair AptosPair) BridgeToAptos {
	return BridgeToAptos{
		aptosPair,
		account.Helper{},
	}
}

func (acc BridgeToAptos) Bridge() {
	constants.Logger.Warn(fmt.Sprintf("Начинаю Aptos Bridge из сети EVM для аккаунта %s", acc.Wallet.PublicKey.String()))
	utils.SendTelegramMessage(fmt.Sprintf("🟨 Начинаю Aptos Bridge из сети EVM для аккаунта %s", acc.Wallet.PublicKey.String()))

	moneyChain, err := acc.FindChainWithMoney("USDC", constants.SETTINGS.Aptos.FromNetworks, acc.Wallet.PublicKey)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Ошибка при поиске сети с балансом USDC")
		utils.SendTelegramMessage("🟥 Ошибка при поиске сети с балансом USDC\n\n", "Ошибка: ", err.Error())
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
		return
	}

	chainFrom := moneyChain.Chain
	toAddress := common.HexToHash("0x" + hex.EncodeToString(acc.Aptos.AuthKey[:]))

	constants.Logger.WithFields(logrus.Fields{
		color.InPurple("Из сети"):        chainFrom,
		color.InPurple("Баланс USDC"):    balance.UiAmount,
		color.InPurple("На аптос адрес"): toAddress,
	}).Trace("Найдена сеть с балансом USDC. Начинаю Aptos бридж")
	utils.SendTelegramMessage("🟪 Найдена сеть с балансом USDC. Начинаю Aptos бридж\n\n",
		"Баланс USDC: ", balance.UiAmount,
		"\nИз сети: ", chainFrom,
		"\nНа аптос адрес: ", toAddress)

	swapArgs := acc.GetAmounts(balance, "aptos")
	amount := swapArgs.Amount

	allowance, err := acc.Allowance(chainFrom, "USDC", "aptos", acc.Wallet)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Ошибка при получении Allowance на USDC")
		utils.SendTelegramMessage("🟥 Ошибка при получении Allowance на USDC\n\n",
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
		}).Warn("Требуется апрув на USDC")
		utils.SendTelegramMessage("🟪 Требуется апрув на USDC\n\n",
			"Amount: ", amount,
			"\nAllowance: ", allowance)

		approve, errApprove := acc.Approve(chainFrom, amount, "USDC", "aptos", acc.Wallet)
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

	swap, err := acc.ToAptosBridgeSwap(chainFrom, swapArgs)
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

	//constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующей активностью", delay))
	//time.Sleep(time.Duration(delay) * time.Second)

}

func (acc BridgeToAptos) ToAptosBridgeSwap(chainFrom string, args account.SwapArgs) (*types.Receipt, error) {
	client := constants.CLIENTS[chainFrom]

	nonce, err := client.PendingNonceAt(context.Background(), acc.Wallet.PublicKey)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["aptos"][chainFrom])
	contractAbi, err := account.ReadAbi(constants.APTOS_BRIDGE_ABI)
	if err != nil {
		return nil, err
	}

	token := common.HexToAddress(constants.CONTRACTS["USDC"][chainFrom])
	toAddress := common.HexToHash("0x" + hex.EncodeToString(acc.Aptos.AuthKey[:]))

	amount := args.Amount

	callParams := struct {
		RefundAddress     common.Address
		ZroPaymentAddress common.Address
	}{
		RefundAddress:     acc.Wallet.PublicKey,
		ZroPaymentAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
	}

	adapterParams := common.Hex2Bytes(acc.GetAdapterParams())

	encodedData, err := contractAbi.Pack("sendToAptos",
		token, toAddress, amount, callParams, adapterParams)
	if err != nil {
		return nil, err
	}

	value, err := acc.GetSwapValue(chainFrom)
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:    &contractAddress,
		From:  acc.Wallet.PublicKey,
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

	canSwap := account.ValidateFee(chainFrom, gasPrice, gasLimit, value)

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

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), acc.Wallet.PrivateKey)
	if err != nil {
		return nil, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		constants.Logger.Debug(gasPrice, gasLimit)
		constants.Logger.Debug(*args.Amount, *args.MinAmount)
		return nil, err
	}

	hash, err := account.WaitMined(client, signedTx)
	if err != nil {
		return nil, err
	}

	return hash, nil

}

func (acc BridgeToAptos) GetSwapValue(chainFrom string) (*big.Int, error) {
	client := constants.CLIENTS[chainFrom]

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["aptos"][chainFrom])

	contractAbi, err := account.ReadAbi(constants.APTOS_BRIDGE_ABI)
	if err != nil {
		return nil, err
	}

	callParams := struct {
		RefundAddress     common.Address
		ZroPaymentAddress common.Address
	}{
		RefundAddress:     acc.Wallet.PublicKey,
		ZroPaymentAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
	}

	adapterParams := common.Hex2Bytes(acc.GetAdapterParams())

	encodedData, err := contractAbi.Pack("quoteForSend", callParams, adapterParams)
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

	res, err := contractAbi.Unpack("quoteForSend", result)
	if err != nil {
		return nil, err
	}

	return res[0].(*big.Int), nil
}

func (acc BridgeToAptos) GetAdapterParams() string {
	gasAmount := big.NewInt(10000)
	nativeForDst := big.NewInt(5000000)

	version := 2

	param1Converted := make([]byte, 2)
	param2Converted := make([]byte, 32)
	param3Converted := make([]byte, 32)
	param4Converted := make([]byte, 32)

	versionBig := big.NewInt(int64(version))
	copy(param1Converted[2-len(versionBig.Bytes()):], versionBig.Bytes())

	gasAmountBytes := gasAmount.Bytes()
	copy(param2Converted[32-len(gasAmountBytes):], gasAmountBytes)

	nativeForDstBytes := nativeForDst.Bytes()
	copy(param3Converted[32-len(nativeForDstBytes):], nativeForDstBytes)

	aptosWalletBytes := acc.Aptos.AuthKey[:]
	copy(param4Converted[32-len(aptosWalletBytes):], aptosWalletBytes)

	adapterParams := append(param1Converted, param2Converted...)
	adapterParams = append(adapterParams, param3Converted...)
	adapterParams = append(adapterParams, param4Converted...)

	params := hex.EncodeToString(adapterParams)

	return params
}

func (acc BridgeToAptos) GetLegacyAdapterParams() string {
	gasLimit := uint(3000000)
	gasPrice := uint64(5000000000)

	// генерируем рандомный байтовый массив длиной 32 байта
	var adapterParams [32]byte
	if _, err := rand.Read(adapterParams[:]); err != nil {
		panic(err)
	}

	// добавляем параметры газа в массив байтов
	adapterParams[16] = byte(gasLimit >> 8)
	adapterParams[17] = byte(gasLimit)
	adapterParams[24] = byte(gasPrice >> 8)
	adapterParams[25] = byte(gasPrice)

	// выводим сгенерированные параметры в виде hex строки
	params := hex.EncodeToString(adapterParams[:])
	return "000200000000000000000000000000000000000000000000000000000000000027100000000000000000000000000000000000000000000000000000000000000000" + params
}
