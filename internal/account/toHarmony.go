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
)

type HarmonyBridgeAccount struct {
	Wallet
	Helper
}

func NewHarmonyBridgeAccount(wallet Wallet) HarmonyBridgeAccount {
	return HarmonyBridgeAccount{
		wallet,
		Helper{},
	}
}

func (acc HarmonyBridgeAccount) Bridge() {
	constants.Logger.Warn(fmt.Sprintf("Начинаю HarmonyBridge для аккаунта %s", acc.PublicKey.String()))
	utils.SendTelegramMessage(fmt.Sprintf("🟨 Начинаю HarmonyBridge для аккаунта %s", acc.PublicKey.String()))

	scan := constants.SCANS["bsc"]

	bridgeSwap, err := acc.HarmonyBridgeSwap("bsc")
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("В сеть"): "harmony",
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

	if bridgeSwap.Status == 0 {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Хэш транзакции"): bridgeSwap.TxHash,
			color.InRed("Ссылка на скан"): scan + bridgeSwap.TxHash.String(),
		}).Error("Транзакция на бридж не прошла")
		utils.SendTelegramMessage("🟥 Транзакция на бридж не прошла\n\n",
			"Ссылка на скан: ", scan+bridgeSwap.TxHash.String())
		errWrite := utils.AppendToFile("errorsOnlyLog.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorsOnlyLog.txt")
		}
		return
	}

	constants.Logger.WithFields(logrus.Fields{
		color.InGreen("Хэш транзакции"): bridgeSwap.TxHash,
		color.InGreen("Ссылка на скан"): scan + bridgeSwap.TxHash.String(),
	}).Info("Транзакция на бридж успешно смайнилась")
	utils.SendTelegramMessage("🟩 Транзакция на бридж успешно смайнилась\n\n",
		"Ссылка на скан: ", scan+bridgeSwap.TxHash.String())
}

func (acc HarmonyBridgeAccount) HarmonyBridgeSwap(chainFrom string) (*types.Receipt, error) {
	client := constants.CLIENTS[chainFrom]

	nonce, err := client.PendingNonceAt(context.Background(), acc.PublicKey)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["harmony"]["bsc"])
	contractAbi, err := utils.ReadAbi(constants.HARMONY_BRIDGE_ABI)
	if err != nil {
		return nil, err
	}

	from := acc.PublicKey
	dstChainId := uint16(116)

	toAddress := acc.PublicKey.Bytes()

	minAmount := constants.SETTINGS.Harmony.NativeMinAmount
	maxAmount := constants.SETTINGS.Harmony.NativeMaxAmount
	amount := utils.GetRandomFloat(minAmount, maxAmount)

	tokenId := utils.ToWei(amount, 18)

	refundAddress := acc.PublicKey
	zroPaymentAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	adapterParams := common.Hex2Bytes(acc.GetAdapterParams())

	encodedData, err := contractAbi.Pack("sendFrom",
		from, dstChainId, toAddress, tokenId, refundAddress, zroPaymentAddress, adapterParams)
	if err != nil {
		return nil, err
	}

	value, err := acc.GetSwapValue(chainFrom, tokenId)
	if err != nil {
		return nil, err
	}

	sum := new(big.Int).Add(tokenId, value)
	//spew.Dump(tokenId, value, sum)

	msg := ethereum.CallMsg{
		To:    &contractAddress,
		From:  acc.PublicKey,
		Value: sum,
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
		sum,
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

func (acc HarmonyBridgeAccount) GetSwapValue(chainFrom string, amount *big.Int) (*big.Int, error) {
	client := constants.CLIENTS[chainFrom]

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["harmony"]["bsc"])

	contractAbi, err := utils.ReadAbi(constants.HARMONY_BRIDGE_ABI)
	if err != nil {
		return nil, err
	}

	dstChainId := uint16(116)
	//toAddress := common.LeftPadBytes(acc.PublicKey.Bytes(), 32)

	//amount := big.NewInt(549)

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

func (acc HarmonyBridgeAccount) GetAdapterParams() string {
	//"0x000200000000000000000000000000000000000000000000000000000000002dc6c00000000000000000000000000000000000000000000000000000000000000000d8ef43f4a095fca7ef6de5d07194029042d38a97"
	return "000100000000000000000000000000000000000000000000000000000000000aae60"
}
