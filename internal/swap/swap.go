package swap

import (
	"filthy/internal/account"
	"filthy/internal/constants"
	"filthy/internal/utils"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"math/big"
)

type SwapAccount struct {
	account.Wallet
	account.Helper
}

func NewSwapAccount(wallet account.Wallet) SwapAccount {
	return SwapAccount{
		wallet,
		account.Helper{},
	}
}

func (acc SwapAccount) Swap() {
	chain := constants.SETTINGS.Swap.Chain
	network := constants.SWAP_CHAINS[chain]

	tokenFrom := constants.SETTINGS.Swap.TokenFrom
	tokenTo := constants.SETTINGS.Swap.TokenTo

	scan := constants.SCANS[chain]

	constants.Logger.WithFields(logrus.Fields{
		color.InYellow("Сеть"):       chain,
		color.InYellow("Токен"):      tokenTo,
		color.InYellow("Покупаю за"): tokenFrom,
	}).Warn("Начинаю свап для аккаунта ", acc.PublicKey.String())
	utils.SendTelegramMessage("🟪 Начинаю свап для аккаунта\n\n",
		"Сеть: ", chain,
		"\nТокен: ", tokenTo,
		"\nПокупаю за: ", tokenFrom)

	tokenAddr := constants.CONTRACTS[tokenFrom][chain]
	toTokenAddr := constants.CONTRACTS[tokenTo][chain]
	slippage := constants.SETTINGS.Accounts.Slippage

	percentOfUscd := constants.SETTINGS.Swap.PercentOfUsdc

	fromNetworks := []string{chain}

	moneyChain, err := acc.FindChainWithMoney(tokenFrom, fromNetworks, acc.PublicKey)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
		}).Error(fmt.Sprintf("Ошибка при поиске %s на аккаунте", tokenFrom))
		utils.SendTelegramMessage(fmt.Sprintf("🟥 Ошибка при поиске %s на аккаунте \n\n", tokenFrom),
			"Ошибка: ", err.Error())
		return
	}

	constants.Logger.WithFields(logrus.Fields{
		color.InPurple(fmt.Sprintf("Баланс %s", tokenFrom)): moneyChain.Balance.UiAmount,
		color.InPurple("Процент от баланса"):                percentOfUscd,
	}).Trace(fmt.Sprintf("Найден баланс %s. Продолжаю", tokenFrom))

	//spew.Dump(moneyChain)

	amount := moneyChain.UiAmount * (percentOfUscd / 100)

	if amount == 0 {
		constants.Logger.Error(fmt.Sprintf("Кол-во %s для свапа равно 0, скипаю", tokenFrom))
		utils.SendTelegramMessage(fmt.Sprintf("Кол-во %s для свапа равно 0, скипаю", tokenFrom))
		return
	}
	//amount := 0.01

	ethBasedClient := NewBasedClient(constants.CLIENTS[chain], acc.Wallet)
	inch, err := BuyTokenByInch(ethBasedClient, tokenAddr, toTokenAddr, amount, int64(slippage), network)
	if err != nil {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Ошибка"): err,
			//"Debug":  swapArgs,
		}).Error("Не удалось отправить транзакцию на свап")
		utils.SendTelegramMessage("🟥 Не удалось отправить транзакцию на свап\n\n",
			"Ошибка: ", err.Error())

		errWrite := utils.AppendToFile("errorSwapWallets.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorSwapWallets.txt")
			return
		}

		return
	}

	if inch.Status == 0 {
		constants.Logger.WithFields(logrus.Fields{
			color.InRed("Хэш транзакции"): inch.TxHash,
			color.InRed("Ссылка на скан"): scan + inch.TxHash.String(),
		}).Error("Транзакция на свап не прошла")
		utils.SendTelegramMessage("🟥 Транзакция на свап не прошла\n\n",
			"Ссылка на скан: ", scan+inch.TxHash.String())

		errWrite := utils.AppendToFile("errorSwapWallets.txt", utils.PrivateKeyToString(acc.PrivateKey))
		if errWrite != nil {
			constants.Logger.Error("Ошибка при записи приватника в файл errorSwapWallets.txt")
			return
		}
		return
	}

	constants.Logger.WithFields(logrus.Fields{
		color.InGreen("Хэш транзакции"): inch.TxHash,
		color.InGreen("Ссылка на скан"): scan + inch.TxHash.String(),
	}).Info("Транзакция на свап успешно смайнилась")

	utils.SendTelegramMessage("🟩 Транзакция на свап успешно смайнилась\n\n",
		"Ссылка на скан: ", scan+inch.TxHash.String())

}

func BuyTokenByInch(ethBasedClient EthBasedClient, tokenAddr string, toTokenAddr string, amount float64, slippage int64, network string) (*types.Receipt, error) {
	usedTokenContractAddr := common.HexToAddress(tokenAddr)
	usedTokenIns := GetTokenInstance(usedTokenContractAddr, ethBasedClient.Client)
	//Logger.Info("Try to Get TokenDecimals")
	usedTokenDecimals := GetTokenDecimals(usedTokenIns)
	amountInWei := EtherToWeiByDecimal(big.NewFloat(amount), int(usedTokenDecimals))
	amountInWeiStr := amountInWei.String()
	//Logger.WithFields(logrus.Fields{"amountInWei": amountInWei, "amountInWeiStr": amountInWeiStr, "usedTokenDecimals": usedTokenDecimals}).Info("amountInWei")

	allowance, err := GetApproveAllowance(tokenAddr, ethBasedClient.Address.String(), network)
	if err != nil {
		return nil, err
	}
	if allowance == "0" {
		approveErr := ApproveTokenByInch(ethBasedClient, tokenAddr, network)
		if approveErr != nil {
			return nil, approveErr
		}
	}

	tx, err := SwapTokenByInch(ethBasedClient, tokenAddr, toTokenAddr, amountInWeiStr, ethBasedClient.Address.String(), slippage, network)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
