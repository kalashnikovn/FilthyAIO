package aptos

import (
	"encoding/hex"
	"filthy/internal/account"
	"filthy/internal/constants"
	"filthy/internal/utils"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/coming-chat/go-aptos/aptostypes"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type BridgeFromAptos struct {
	AptosPair
	account.Helper
}

func NewBridgeFromAptos(aptosPair AptosPair) BridgeFromAptos {
	return BridgeFromAptos{
		aptosPair,
		account.Helper{},
	}
}

func (acc BridgeFromAptos) Bridge() {
	fromAddress := "0x" + hex.EncodeToString(acc.Aptos.AuthKey[:])
	constants.Logger.Warn(fmt.Sprintf("Начинаю Aptos Bridge из сети Aptos для аккаунта %s", fromAddress))
	utils.SendTelegramMessage(fmt.Sprintf("Начинаю Aptos Bridge из сети Aptos для аккаунта %s", fromAddress))

	amountUSDC, err := acc.GetUSDCBalance()
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Ошибка при получении баланса USDC. Возможно проблема в ноде")
		utils.SendTelegramMessage("🟥 Ошибка при получении баланса USDC. Возможно проблема в ноде\n\n",
			"Ошибка: ", err.Error())
		return
	}

	if amountUSDC == "0" {
		constants.Logger.Error("Баланс USDC на аккаунте в сети Aptos равен нулю. Скипаю...")
		utils.SendTelegramMessage("Баланс USDC на аккаунте в сети Aptos равен нулю. Скипаю...")
		return
	}

	constants.Logger.WithFields(logrus.Fields{
		color.InPurple("Из сети"):                "Aptos",
		color.InPurple("В сеть"):                 "Avalanche",
		color.InPurple("Баланс USDC (*big.Int)"): amountUSDC,
		color.InPurple("На EVM адрес"):           acc.Wallet.PublicKey.String(),
	}).Trace("Найден баланс USDC. Начинаю бридж из Aptos в Avalanche")
	utils.SendTelegramMessage("🟪 Найден баланс USDC. Начинаю бридж из Aptos в Avalanche\n\n",
		"Баланс USDC (*big.Int): ", amountUSDC,
		"\nИз сети: ", "Aptos",
		"\nВ сеть: ", "Avalanche",
		"\nНа EVM адрес: ", acc.Wallet.PublicKey.String())

	constants.Logger.Warn("Отравляю транзакцию на мост...")

	swap, err := acc.BridgeFromAptosSwap(amountUSDC)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Не удалось отправить транзакцию на бридж")
		utils.SendTelegramMessage("🟥 Не удалось отправить транзакцию на бридж\n\n",
			"Ошибка: ", err.Error())
		return
	}

	scan := constants.SCANS["aptos"]

	constants.Logger.WithFields(logrus.Fields{
		color.InGreen("Хэш транзакции"): swap.Hash,
		color.InGreen("Ссылка на скан"): scan + swap.Hash + "/userTxnOverview?network=mainnet",
	}).Info("Транзакция из сети Aptos в Avalanche успешно отправлена")
	utils.SendTelegramMessage("🟩 Транзакция из сети Aptos в Avalanche успешно отправлена\n\n",
		"Ссылка на скан: ", scan+swap.Hash+"/userTxnOverview?network=mainnet")

	delay := acc.GetRandomActivityDelay()

	constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующей активностью", delay))
	utils.SendTelegramMessage(fmt.Sprintf("Сплю %d секунд перед следующей активностью", delay))

	time.Sleep(time.Duration(delay) * time.Second)

}

func (acc BridgeFromAptos) BridgeFromAptosSwap(amountUSDC string) (*aptostypes.Transaction, error) {
	fromAddress := "0x" + hex.EncodeToString(acc.Aptos.AuthKey[:])
	accountData, err := acc.AptosClient.GetAccount(fromAddress)
	if err != nil {
		return nil, err
	}

	sequenceNumber := accountData.SequenceNumber

	ledgerInfo, err := acc.AptosClient.LedgerInfo()
	if err != nil {
		return nil, err
	}

	toAddressEVM := account.СonvertAddress(acc.Wallet.PublicKey.String())
	//amountUSDC, err := acc.GetUSDCBalance()
	//if err != nil {
	//	return nil, err
	//}

	randomFee := acc.GenerateRandomValue(3_513_251, 5_000_000)

	payload := &aptostypes.Payload{
		Type:     "entry_function_payload",
		Function: "0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::coin_bridge::send_coin_from",
		TypeArguments: []string{
			"0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::asset::USDC",
		},
		Arguments: []interface{}{
			"106",
			toAddressEVM,
			amountUSDC,
			randomFee,
			"0",
			false,
			"0x000100000000000249f0",
			"0x",
		},
	}

	gasPrice, err := acc.AptosClient.EstimateGasPrice()
	if err != nil {
		return nil, err
	}

	transaction := &aptostypes.Transaction{
		Sender:                  fromAddress,
		SequenceNumber:          sequenceNumber,
		MaxGasAmount:            2000,
		GasUnitPrice:            gasPrice,
		Payload:                 payload,
		ExpirationTimestampSecs: ledgerInfo.LedgerTimestamp + 10000000, // 10 minutes timeout
	}

	signingMessage, err := acc.AptosClient.CreateTransactionSigningMessage(transaction)
	if err != nil {
		return nil, err
	}

	signatureData := acc.Aptos.Sign(signingMessage, "")
	signatureHex := "0x" + hex.EncodeToString(signatureData)
	publicKey := "0x" + hex.EncodeToString(acc.Aptos.PublicKey)
	transaction.Signature = &aptostypes.Signature{
		Type:      "ed25519_signature",
		PublicKey: publicKey,
		Signature: signatureHex,
	}

	newTx, err := acc.AptosClient.SubmitTransaction(transaction)
	if err != nil {
		return nil, err
	}

	return newTx, nil
}

func (acc BridgeFromAptos) GetUSDCBalance() (string, error) {
	fromAddress := "0x" + hex.EncodeToString(acc.Aptos.AuthKey[:])
	of, err := acc.AptosClient.BalanceOf(fromAddress, "0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::asset::USDC")
	if err != nil {
		return "0", err
	}

	return of.String(), nil
}

func (acc BridgeFromAptos) GenerateRandomValue(min, max int) string {
	// Инициализация генератора случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Генерация случайного числа
	randomNumber := rand.Intn(max-min+1) + min

	// Преобразование в строку
	randomString := fmt.Sprintf("%d", randomNumber)

	return randomString
}
