package staking

import (
	"context"
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
	"time"
)

type StakingAccount struct {
	account.Wallet
	account.Helper
}

func NewStakingAccount(wallet account.Wallet) StakingAccount {
	return StakingAccount{
		wallet,
		account.Helper{},
	}
}

func (acc StakingAccount) Lock() {
	chainFrom := constants.SETTINGS.Staking.FromNetwork
	fromNetworks := []string{chainFrom}
	moneyChain, err := acc.FindChainWithMoney("STG", fromNetworks, acc.PublicKey)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Ошибка при поиске сети с балансом STG")
		return
	}

	balance := moneyChain.Balance
	bigBalance := balance.BigAmount

	if balance.UiAmount == 0 {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Сеть"):       moneyChain.Chain,
			color.InRed("Баланс STG"): balance.UiAmount,
		}).Error("Недостаточный баланс STG на аккаунте")
		return
	}
	scan := constants.SCANS[chainFrom]

	constants.Logger.WithFields(logrus.Fields{
		color.InPurple("Из сети"):    chainFrom,
		color.InPurple("Баланс STG"): balance.UiAmount,
	}).Trace("Найдена сеть с балансом STG. Начинаю Stargate стейкинг")

	allowance, err := acc.Allowance(chainFrom, "STG", "staking", acc.Wallet)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error("Ошибка при получении Allowance на STG")
		return
	}

	cmp := bigBalance.Cmp(allowance)
	if cmp == 1 {
		//fmt.Println("need approve")
		constants.Logger.WithFields(logrus.Fields{
			color.InYellow("Amount"):    bigBalance,
			color.InYellow("Allowance"): allowance,
		}).Warn("Требуется апрув на STG")

		approve, errApprove := acc.Approve(chainFrom, bigBalance, "STG", "staking", acc.Wallet)
		if errApprove != nil {
			constants.Logger.WithFields(logrus.Fields{
				color.InRed("Ошибка"): errApprove,
			}).Error("Не удалось сделать апрув")
			return
		}

		if approve.Status == 0 {
			constants.Logger.WithFields(logrus.Fields{
				color.InRed("Хэш транзакции"): approve.TxHash,
				color.InRed("Ссылка на скан"): scan + approve.TxHash.String(),
			}).Error("Транзакция на апрув не прошла")
			return
		}

		constants.Logger.WithFields(logrus.Fields{
			color.InGreen("Хэш транзакции"): approve.TxHash,
			color.InGreen("Ссылка на скан"): scan + approve.TxHash.String(),
		}).Info("Транзакция на апрув успешно смайнилась")

		delay := acc.GetRandomActivityDelay()

		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующей транзакцией", delay))
		time.Sleep(time.Duration(delay) * time.Second)
	}

	constants.Logger.Warn("Отравляю транзакцию на стейкинг...")

	swap, err := acc.CreateLock(chainFrom, bigBalance)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
			//"Debug":  swapArgs,
		}).Error("Не удалось отправить транзакцию на стейкинг")
		return
	}

	if swap.Status == 0 {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Хэш транзакции"): swap.TxHash,
			color.InRed("Ссылка на скан"): scan + swap.TxHash.String(),
		}).Error("Транзакция на стейкинг не прошла")
		return
	}

	constants.Logger.WithFields(logrus.Fields{
		color.InGreen("Хэш транзакции"): swap.TxHash,
		color.InGreen("Ссылка на скан"): scan + swap.TxHash.String(),
	}).Info("Транзакция на стейкинг успешно смайнилась")

}

func (acc StakingAccount) CreateLock(chainFrom string, amount *big.Int) (*types.Receipt, error) {
	client := constants.CLIENTS[chainFrom]

	nonce, err := client.PendingNonceAt(context.Background(), acc.PublicKey)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(constants.STARGATE_LOCK_CONTRACTS[chainFrom])
	contractAbi, err := utils.ReadAbi(constants.STARGATE_LOCK_ABI)
	if err != nil {
		return nil, err
	}

	currentTimestamp := time.Now().Unix()
	//timestamp := acc.AddMonthsToTimestamp(currentTimestamp)
	lockMonths := constants.SETTINGS.Staking.LockPeriod
	timestamp := acc.AddMonthsToTimestamp(currentTimestamp, lockMonths)

	encodedData, err := contractAbi.Pack("create_lock",
		amount, timestamp)
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:    &contractAddress,
		From:  acc.PublicKey,
		Value: big.NewInt(0),
		Data:  encodedData,
	}

	//gasPrice, err := client.SuggestGasPrice(context.Background())
	gasPrice, err := utils.GetGasPrice(client, chainFrom)
	if err != nil {
		return nil, err
	}

	gasLimit, err := utils.GetGasLimit(client, msg)
	//spew.Dump(gasPrice, gasLimit)
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(
		nonce,
		contractAddress,
		big.NewInt(0),
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

func (acc StakingAccount) AddMonthsToTimestamp(timestamp int64, monthsToAdd int) *big.Int {
	// Преобразуем Unix-таймштамп в объект времени
	t := time.Unix(timestamp, 0)

	// Добавляем 36 месяцев
	newTime := t.AddDate(0, monthsToAdd, 0)

	// Преобразуем новое время в Unix-таймштамп
	newTimestamp := newTime.Unix()

	// Создаем объект big.Int и устанавливаем в него новый таймштамп
	timestampBigInt := big.NewInt(newTimestamp)

	return timestampBigInt
}
