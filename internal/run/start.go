package run

import (
	"filthy/internal/account"
	"filthy/internal/aptos"
	"filthy/internal/constants"
	"filthy/internal/refuel"
	"filthy/internal/staking"
	"filthy/internal/swap"
	"filthy/internal/utils"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/TwiN/go-color"
	"time"
)

func StartRandom(wallets []account.Wallet) {
	bridges := constants.SETTINGS.Accounts.RandomBridges
	PrintRandomInfo(bridges)
	PrintWallets(wallets)

	cont := false
	prompt := &survey.Confirm{
		Message: "Запускаем?",
	}
	survey.AskOne(prompt, &cont)

	if cont == false {
		return
	}

	cycles := constants.SETTINGS.Accounts.Cycles

	walletsLen := len(wallets)
	bridgesLen := len(bridges)

	constants.Logger.Warn("Кол-во аккаунтов: ", walletsLen)
	constants.Logger.Warn("Общее кол-во активностей в одном цикле: ", walletsLen*bridgesLen)
	println()

	for cycle := 0; cycle < cycles; cycle += 1 {
		constants.Logger.Warn("Запускаю цикл номер ", cycle+1, ". Всего циклов ", cycles)
		utils.SendTelegramMessage("Запускаю цикл номер ", cycle+1, ". Всего циклов ", cycles)

		println("\n")

		paths := account.GetRandomPaths(bridges, wallets)

		doneActivity := 0

		for _, acc := range paths {
			for _, bridge := range acc {
				println("-------------------------------------------------------------------------------------------------------------")
				bridge.Bridge()
				doneActivity += 1
				println("-------------------------------------------------------------------------------------------------------------")
				println("\n")

				percent := (float32(doneActivity) / (float32(walletsLen) * float32(bridgesLen))) * 100

				println(color.InRed("-------------------------------------------------------------------------------------------------------------"))
				println(color.InYellow(fmt.Sprintf("Выполнено %.1f%% активностей в текущем цикле", percent)))
				utils.SendTelegramMessage(fmt.Sprintf("Выполнено %.1f%% активностей в текущем цикле", percent))
				println(color.InRed("-------------------------------------------------------------------------------------------------------------"))
				println("\n")
			}

			delay := GetRandomAccsDelay()
			constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
			utils.SendTelegramMessage(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
			time.Sleep(time.Duration(delay) * time.Second)
			println("\n")
		}

		delay := constants.SETTINGS.Accounts.CycleDelay
		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующим циклом", delay))
		utils.SendTelegramMessage(fmt.Sprintf("Сплю %d секунд перед следующим циклом", delay))
		time.Sleep(time.Duration(delay) * time.Second)
		println("\n")
	}

}

func StartSwap(wallets []account.Wallet) {
	PrintSwapInfo()
	PrintWallets(wallets)

	cont := false
	prompt := &survey.Confirm{
		Message: "Запускаем?",
	}
	survey.AskOne(prompt, &cont)

	if cont == false {
		return
	}

	constants.Logger.Warn("Запускаю свап на аккаунтах ")
	println("\n")

	for _, wallet := range wallets {
		acc := swap.NewSwapAccount(wallet)
		println("-------------------------------------------------------------------------------------------------------------")
		acc.Swap()
		println("-------------------------------------------------------------------------------------------------------------")
		println()

		delay := GetRandomAccsDelay()
		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		time.Sleep(time.Duration(delay) * time.Second)
		println("\n")
	}

}

func StartRefuel(wallets []account.Wallet) {
	PrintRefuelInfo()
	PrintWallets(wallets)

	cont := false
	prompt := &survey.Confirm{
		Message: "Запускаем?",
	}
	survey.AskOne(prompt, &cont)

	if cont == false {
		return
	}

	chains := constants.SETTINGS.Refuel.ToNetworks

	constants.Logger.Warn("Запускаю рефуел на аккаунтах")
	println("\n")

	for _, wallet := range wallets {
		acc := refuel.NewRefuelAccount(wallet)
		println("-------------------------------------------------------------------------------------------------------------")
		acc.Refuel(chains)
		println("-------------------------------------------------------------------------------------------------------------")
		println()

		delay := GetRandomAccsDelay()
		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		time.Sleep(time.Duration(delay) * time.Second)
		println("\n")
	}

}

func StartTestnet(wallets []account.Wallet) {
	PrintTestnetInfo()
	PrintWallets(wallets)

	cont := false
	prompt := &survey.Confirm{
		Message: "Запускаем?",
	}
	survey.AskOne(prompt, &cont)

	if cont == false {
		return
	}

	constants.Logger.Warn("Запускаю TestnetBridge на аккаунтах")
	println("\n")

	for _, wallet := range wallets {
		acc := account.NewTestnetBridgeAccount(wallet)
		println("-------------------------------------------------------------------------------------------------------------")
		acc.Bridge()
		println("-------------------------------------------------------------------------------------------------------------")
		println()

		delay := GetRandomAccsDelay()
		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		utils.SendTelegramMessage(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		time.Sleep(time.Duration(delay) * time.Second)
		println("\n")
	}

}

func StartHarmony(wallets []account.Wallet) {
	PrintHarmonyInfo()
	PrintWallets(wallets)

	cont := false
	prompt := &survey.Confirm{
		Message: "Запускаем?",
	}
	survey.AskOne(prompt, &cont)

	if cont == false {
		return
	}

	constants.Logger.Warn("Запускаю HarmonyBridge на аккаунтах")
	println("\n")

	for _, wallet := range wallets {
		acc := account.NewHarmonyBridgeAccount(wallet)
		println("-------------------------------------------------------------------------------------------------------------")
		acc.Bridge()
		println("-------------------------------------------------------------------------------------------------------------")
		println()

		delay := GetRandomAccsDelay()
		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		utils.SendTelegramMessage(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		time.Sleep(time.Duration(delay) * time.Second)
		println("\n")
	}

}

func StartCoreDao(wallets []account.Wallet) {
	PrintCoreInfo("bsc")
	PrintWallets(wallets)

	cont := false
	prompt := &survey.Confirm{
		Message: "Запускаем?",
	}
	survey.AskOne(prompt, &cont)

	if cont == false {
		return
	}

	constants.Logger.Warn("Запускаю CoreDao Bridge на аккаунтах")
	println("\n")

	for _, wallet := range wallets {
		acc := account.NewCoreBridgeAccount(wallet)
		println("-------------------------------------------------------------------------------------------------------------")
		acc.Bridge()
		println("-------------------------------------------------------------------------------------------------------------")
		println()

		delay := GetRandomAccsDelay()
		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		utils.SendTelegramMessage(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		time.Sleep(time.Duration(delay) * time.Second)
		println("\n")
	}

}

func StartFromCoreDao(wallets []account.Wallet) {
	PrintCoreInfo("Core Dao")
	PrintWallets(wallets)

	cont := false
	prompt := &survey.Confirm{
		Message: "Запускаем?",
	}
	survey.AskOne(prompt, &cont)

	if cont == false {
		return
	}

	constants.Logger.Warn("Запускаю From CoreDao Bridge на аккаунтах")
	println("\n")

	for _, wallet := range wallets {
		acc := account.NewFromCoreBridgeAccount(wallet)
		println("-------------------------------------------------------------------------------------------------------------")
		acc.Bridge()
		println("-------------------------------------------------------------------------------------------------------------")
		println()

		delay := GetRandomAccsDelay()
		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		utils.SendTelegramMessage(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		time.Sleep(time.Duration(delay) * time.Second)
		println("\n")
	}

}

func StartBridgeToAptos(wallets []aptos.AptosPair) {
	PrintAptosInfo()
	PrintPairsWallets(wallets)

	cont := false
	prompt := &survey.Confirm{
		Message: "Запускаем?",
	}
	survey.AskOne(prompt, &cont)

	if cont == false {
		return
	}

	constants.Logger.Warn("Запускаю AptosBridge на аккаунтах")
	println("\n")

	for _, wallet := range wallets {
		acc := aptos.NewBridgeToAptos(wallet)
		println("-------------------------------------------------------------------------------------------------------------")
		acc.Bridge()
		println("-------------------------------------------------------------------------------------------------------------")
		println()

		delay := GetRandomAccsDelay()
		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		utils.SendTelegramMessage(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		time.Sleep(time.Duration(delay) * time.Second)
		println("\n")
	}
}

func StartBridgeFromAptos(wallets []aptos.AptosPair) {
	PrintFromAptosInfo()
	PrintFromPairsWallets(wallets)

	cont := false
	prompt := &survey.Confirm{
		Message: "Запускаем?",
	}
	survey.AskOne(prompt, &cont)

	if cont == false {
		return
	}

	constants.Logger.Warn("Запускаю From Aptos Bridge на аккаунтах")
	println("\n")

	for _, wallet := range wallets {
		acc := aptos.NewBridgeFromAptos(wallet)
		println("-------------------------------------------------------------------------------------------------------------")
		acc.Bridge()
		println("-------------------------------------------------------------------------------------------------------------")
		println()

		delay := GetRandomAccsDelay()
		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		utils.SendTelegramMessage(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		time.Sleep(time.Duration(delay) * time.Second)
		println("\n")
	}
}

func StartStaking(wallets []account.Wallet) {
	PrintStakingInfo()
	PrintWallets(wallets)

	cont := false
	prompt := &survey.Confirm{
		Message: "Запускаем?",
	}
	survey.AskOne(prompt, &cont)

	if cont == false {
		return
	}

	constants.Logger.Warn("Запускаю стейкинг на аккаунтах ")
	println("\n")

	for _, wallet := range wallets {
		acc := staking.NewStakingAccount(wallet)
		println("-------------------------------------------------------------------------------------------------------------")
		acc.Lock()
		println("-------------------------------------------------------------------------------------------------------------")
		println()

		delay := GetRandomAccsDelay()
		constants.Logger.Trace(fmt.Sprintf("Сплю %d секунд перед следующим аккаунтом", delay))
		time.Sleep(time.Duration(delay) * time.Second)
		println("\n")
	}

}
