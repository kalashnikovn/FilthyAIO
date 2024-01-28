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

type StargateAccount struct {
	Wallet
	Helper
}

func NewStargateAccount(wallet Wallet) StargateAccount {
	return StargateAccount{
		wallet,
		Helper{},
	}
}

func (acc StargateAccount) Bridge() {
	constants.Logger.Warn(fmt.Sprintf("Начинаю Stargate Bridge для аккаунта %s", acc.PublicKey.String()))
	utils.SendTelegramMessage(fmt.Sprintf("🟨 Начинаю Stargate Bridge для аккаунта %s", acc.PublicKey.String()))

	moneyChain, err := acc.FindChainWithMoney("USDC", constants.SETTINGS.Stargate.FromNetworks, acc.PublicKey)
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
	randomTo, err := utils.RandomChain(constants.SETTINGS.Stargate.ToNetworks, chainFrom)
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

	//spew.Dump(randomTo)

	constants.Logger.WithFields(logrus.Fields{
		color.InPurple("Из сети"):     chainFrom,
		color.InPurple("В сеть"):      randomTo,
		color.InPurple("Баланс USDC"): balance.UiAmount,
	}).Trace("Найдена сеть с балансом USDC. Начинаю Stargate бридж")
	utils.SendTelegramMessage("🟪 Найдена сеть с балансом USDC. Начинаю Stargate бридж\n\n",
		"Баланс USDC: ", balance.UiAmount,
		"\nИз сети: ", chainFrom,
		"\nВ сеть: ", randomTo)

	swapArgs := acc.GetAmounts(balance, "stargate")

	//spew.Dump(swapArgs)
	amount := swapArgs.Amount

	allowance, err := acc.Allowance(chainFrom, "USDC", "stargate", acc.Wallet)
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
		//fmt.Println("need approve")
		constants.Logger.WithFields(logrus.Fields{
			color.InYellow("Amount"):    amount,
			color.InYellow("Allowance"): allowance,
		}).Warn("Требуется апрув на USDC")
		utils.SendTelegramMessage("🟪 Требуется апрув на USDC\n\n",
			"Amount: ", amount,
			"\nAllowance: ", allowance)

		approve, errApprove := acc.Approve(chainFrom, amount, "USDC", "stargate", acc.Wallet)
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

	swap, err := acc.StargateSwap(chainFrom, randomTo, swapArgs)
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

func (acc StargateAccount) StargateSwap(chainFrom, chainTo string, args SwapArgs) (*types.Receipt, error) {
	client := constants.CLIENTS[chainFrom]

	nonce, err := client.PendingNonceAt(context.Background(), acc.PublicKey)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["stargate"][chainFrom])

	contractAbi, err := ReadAbi(constants.STARGATE_ABI)
	if err != nil {
		return nil, err
	}

	dstChainId := constants.STARGATE_CHAIN_ID[chainTo]
	srcPoolId := big.NewInt(constants.STARGATE_POOL_ID[chainFrom])
	dstPoolId := big.NewInt(constants.STARGATE_POOL_ID[chainTo])
	refundAddress := acc.PublicKey
	amountLD := args.Amount
	minAmountLD := args.MinAmount
	nativeAddr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	lzTxParams := struct {
		DstGasForCall   *big.Int
		DstNativeAmount *big.Int
		DstNativeAddr   []byte
	}{
		DstGasForCall:   big.NewInt(0),
		DstNativeAmount: big.NewInt(0),
		DstNativeAddr:   nativeAddr.Bytes(),
	}

	to := acc.PublicKey.Bytes()
	transferAndCallPayload := []byte{}

	encodedData, err := contractAbi.Pack("swap",
		dstChainId, srcPoolId, dstPoolId, refundAddress, amountLD, minAmountLD, lzTxParams, to, transferAndCallPayload)
	if err != nil {
		return nil, err
	}

	value, err := acc.GetSwapValue(chainFrom, chainTo)
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

	//gasFeeCap, err := utils.MultiplyBigInt(gasPrice, 2.0)
	//if err != nil {
	//	return nil, err
	//}
	//
	//tx := types.NewTx(&types.DynamicFeeTx{
	//	ChainID:   chainID,
	//	Nonce:     nonce,
	//	GasFeeCap: gasFeeCap,
	//	GasTipCap: gasPrice,
	//	Gas:       gasLimit,
	//	To:        &contractAddress,
	//	Value:     value,
	//	Data:      encodedData,
	//})

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), acc.PrivateKey)
	if err != nil {
		return nil, err
	}

	//fmt.Println("Bridge", signedTx.Hash())

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

func (acc StargateAccount) GetSwapValue(chainFrom, chainTo string) (*big.Int, error) {
	client := constants.CLIENTS[chainFrom]

	contractAddress := common.HexToAddress(constants.BRIDGE_CONTRACTS["stargate"][chainFrom])

	contractAbi, err := ReadAbi(constants.STARGATE_ABI)
	if err != nil {
		return nil, err
	}

	dstChainId := constants.STARGATE_CHAIN_ID[chainTo]
	functionType := uint8(1)
	toAddress := common.HexToAddress("0x0000000000000000000000000000000000001010").Bytes()
	transferAndCallPayload := []byte{}
	lzTxParams := struct {
		DstGasForCall   *big.Int
		DstNativeAmount *big.Int
		DstNativeAddr   []byte
	}{
		DstGasForCall:   big.NewInt(0),
		DstNativeAmount: big.NewInt(0),
		DstNativeAddr:   common.HexToAddress("0x0000000000000000000000000000000000000001").Bytes(),
	}

	encodedData, err := contractAbi.Pack("quoteLayerZeroFee", dstChainId, functionType, toAddress, transferAndCallPayload, lzTxParams)
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
