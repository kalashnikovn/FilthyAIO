package account

import (
	"math/rand"
	"time"
)

type BridgeAccount interface {
	Bridge()
}

func NewBridgeAccount(bridge string, wallet Wallet) BridgeAccount {
	switch bridge {
	case "stargate":
		return NewStargateAccount(wallet)
	case "btcb":
		return NewBtcBridgeAccount(wallet)
	case "eurb":
		return NewEurBridgeAccount(wallet)
	case "woofi":
		return NewWoofiBridgeAccount(wallet)
	case "testnet":
		return NewTestnetBridgeAccount(wallet)
	case "harmony":
		return NewHarmonyBridgeAccount(wallet)
	}

	return nil
}

func GetRandomPathForAcc(bridges []string, wallet Wallet) []BridgeAccount {
	var bridgeAccounts []BridgeAccount

	for _, bridge := range bridges {
		bridgeAccounts = append(bridgeAccounts, NewBridgeAccount(bridge, wallet))
	}
	shuffle(bridgeAccounts)

	return bridgeAccounts

}

func GetRandomPaths(bridges []string, wallets []Wallet) [][]BridgeAccount {
	var accs [][]BridgeAccount

	for _, wallet := range wallets {
		accs = append(accs, GetRandomPathForAcc(bridges, wallet))
	}
	shuffle2(accs)

	return accs
}

func shuffle(arr []BridgeAccount) {
	if len(arr) <= 1 {
		return
	}
	rand.Seed(time.Now().UnixNano()) // Инициализируем генератор случайных чисел
	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)           // Генерируем случайный индекс от 0 до i включительно
		arr[i], arr[j] = arr[j], arr[i] // Меняем значения местами
	}
}

func shuffle2(arr [][]BridgeAccount) {
	if len(arr) <= 1 {
		return
	}
	rand.Seed(time.Now().UnixNano()) // Инициализируем генератор случайных чисел
	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)           // Генерируем случайный индекс от 0 до i включительно
		arr[i], arr[j] = arr[j], arr[i] // Меняем значения местами
	}
}
