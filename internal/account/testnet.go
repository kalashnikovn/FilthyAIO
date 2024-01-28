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

type TestnetBridgeAccount struct {
	Wallet
	Helper
}

func NewTestnetBridgeAccount(wallet Wallet) TestnetBridgeAccount {
	return TestnetBridgeAccount{
		wallet,
		Helper{},
	}
}

func (acc TestnetBridgeAccount) Bridge() {
	constants.Logger.Warn(fmt.Sprintf("Начинаю TestnetBridge для аккаунта %s", acc.PublicKey.String()))
	utils.SendTelegramMessage(fmt.Sprintf("🟨 Начинаю Testnet Bridge для аккаунта %s", acc.PublicKey.String()))

	chainFrom := constants.SETTINGS.Testnet.FromNetwork
	scan := constants.SCANS[chainFrom]

	bridgeSwap, err := acc.TestnetBridgeSwap(chainFrom)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("В сеть"): "goerli",
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

func (acc TestnetBridgeAccount) TestnetBridgeSwap(chainFrom string) (*types.Receipt, error) {
	client := constants.CLIENTS[chainFrom]

	nonce, err := client.PendingNonceAt(context.Background(), acc.PublicKey)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["testnet"][chainFrom])
	contractAbi, err := ReadAbi(constants.TESTNET_BRIDGE_ABI)
	if err != nil {
		return nil, err
	}

	//amountIn := ToWei(constants.SETTINGS.Testnet.AmountETH)

	minAmount := constants.SETTINGS.Testnet.NativeMinAmount
	maxAmount := constants.SETTINGS.Testnet.NativeMaxAmount
	amount := utils.GetRandomFloat(minAmount, maxAmount)
	amountIn := ToWei(amount)

	amountOutMin, err := acc.GetAmountOut(chainFrom, amountIn)
	if err != nil {
		return nil, err
	}

	dstChainId := constants.STARGATE_CHAIN_ID["goerli"]
	to := acc.PublicKey
	refundAddress := acc.PublicKey
	zroPaymentAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	adapterParams := []byte{}

	encodedData, err := contractAbi.Pack("swapAndBridge",
		amountIn, amountOutMin, dstChainId, to, refundAddress, zroPaymentAddress, adapterParams)
	if err != nil {
		return nil, err
	}

	value, err := acc.GetSwapValue(chainFrom, amountIn)
	if err != nil {
		return nil, err
	}
	value.Add(value, amountIn)

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

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		constants.Logger.Debug(gasPrice, gasLimit)
		constants.Logger.Debug(value)
		return nil, err
	}

	hash, err := WaitMined(client, signedTx)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (acc TestnetBridgeAccount) GetAmountOut(chainFrom string, amountIn *big.Int) (*big.Int, error) {
	client := constants.CLIENTS[chainFrom]

	contractAddress := common.HexToAddress(constants.UNISWAP_ADDRESS)
	contractAbi, err := ReadAbi(constants.UNISWAP_ABI)
	if err != nil {
		return nil, err
	}

	tokenIn := common.HexToAddress(constants.CONTRACTS["WETH"][chainFrom])
	tokenOut := common.HexToAddress(constants.CONTRACTS["GETH"][chainFrom])
	fee := big.NewInt(3000)

	sqrtPriceLimitX96 := big.NewInt(0)

	encodedData, err := contractAbi.Pack("quoteExactInputSingle",
		tokenIn, tokenOut, fee, amountIn, sqrtPriceLimitX96)
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

	res, err := contractAbi.Unpack("quoteExactInputSingle", result)
	if err != nil {
		return nil, err
	}

	return res[0].(*big.Int), nil

}

func (acc TestnetBridgeAccount) GetSwapValue(chainFrom string, amountIn *big.Int) (*big.Int, error) {
	client := constants.CLIENTS[chainFrom]

	contractAddress := common.HexToAddress(constants.CONTRACTS["GETH"][chainFrom])

	contractAbi, err := ReadAbi(constants.GETH_ABI)
	if err != nil {
		return nil, err
	}

	dstChainId := constants.STARGATE_CHAIN_ID["goerli"]

	toAddress := acc.PublicKey.Bytes()

	adapterParams := []byte{}

	encodedData, err := contractAbi.Pack("estimateSendFee",
		dstChainId, toAddress, amountIn, true, adapterParams)
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
