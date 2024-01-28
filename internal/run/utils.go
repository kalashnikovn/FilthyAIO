package run

import (
	"encoding/hex"
	"filthy/internal/account"
	"filthy/internal/aptos"
	"filthy/internal/constants"
	"github.com/TwiN/go-color"
	"github.com/ethereum/go-ethereum/common"
	"math/rand"
	"time"
)

func PrintWallets(wallets []account.Wallet) {
	println(color.InRed("Imported wallets"))
	for _, wallet := range wallets {
		println(color.InCyan("  " + wallet.PublicKey.String()))
	}
	println()
}

func PrintPairsWallets(wallets []aptos.AptosPair) {
	println(color.InRed("Imported wallets"))
	for _, wallet := range wallets {
		toAddress := common.HexToHash("0x" + hex.EncodeToString(wallet.Aptos.AuthKey[:]))
		println("  EVM "+color.InCyan(wallet.Wallet.PublicKey.String()), " to APTOS ", color.InGreen(toAddress))
	}
	println()

}

func PrintFromPairsWallets(wallets []aptos.AptosPair) {
	println(color.InRed("Imported wallets"))
	for _, wallet := range wallets {
		toAddress := common.HexToHash("0x" + hex.EncodeToString(wallet.Aptos.AuthKey[:]))
		println("  APTOS ", color.InGreen(toAddress), " to EVM "+color.InCyan(wallet.Wallet.PublicKey.String()))
	}
	println()

}

func GetRandomAccsDelay() int {
	activity := constants.SETTINGS.Accounts.DelayBetweenAccounts
	min := activity[0]
	max := activity[1]
	rand.Seed(time.Now().UnixNano())
	intn := rand.Intn(max-min+1) + min
	return intn
}

func GetRandomCycleDelay() int {
	max := constants.SETTINGS.Accounts.CycleDelay
	min := 1
	rand.Seed(time.Now().UnixNano())
	intn := rand.Intn(max-min+1) + min
	return intn
}
