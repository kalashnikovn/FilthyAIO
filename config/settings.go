package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Settings struct {
	AuthKey  string `json:"authKey"`
	Telegram struct {
		ChatId   int64  `json:"chatId"`
		BotToken string `json:"botToken"`
	} `json:"telegram"`
	Accounts struct {
		DelayBetweenAccounts []int    `json:"delayBetweenAccounts"`
		DelayActivity        []int    `json:"delayActivity"`
		CycleDelay           int      `json:"cycleDelay"`
		Cycles               int      `json:"cycles"`
		Slippage             float64  `json:"slippage"`
		MinUsdcBalance       float64  `json:"minUsdcBalance"`
		RandomBridges        []string `json:"randomBridges"`
		BscGwei              int64    `json:"bscGwei"`
	} `json:"accounts"`
	Fee      map[string]float64 `json:"fee"`
	Stargate BridgeSettings     `json:"stargate"`
	BtcB     BridgeSettings     `json:"btc.b"`
	WooFi    BridgeSettings     `json:"woofi"`
	EurB     BridgeSettings     `json:"eur.b"`
	Aptos    BridgeSettings     `json:"aptos"`
	Harmony  struct {
		NativeMinAmount float64 `json:"nativeMinAmount"`
		NativeMaxAmount float64 `json:"nativeMaxAmount"`
	} `json:"harmony"`
	CoreDao struct {
		BridgeAmountPercent float64 `json:"bridgeAmountPercent"`
	} `json:"coreDao"`
	Testnet struct {
		FromNetwork     string  `json:"fromNetwork"`
		NativeMinAmount float64 `json:"nativeMinAmount"`
		NativeMaxAmount float64 `json:"nativeMaxAmount"`
	} `json:"testnet"`
	Swap struct {
		Chain         string  `json:"chain"`
		TokenFrom     string  `json:"tokenFrom"`
		TokenTo       string  `json:"tokenTo"`
		PercentOfUsdc float64 `json:"percentOfTokenBalance"`
	} `json:"swap"`
	Staking struct {
		FromNetwork string `json:"fromNetwork"`
		LockPeriod  int    `json:"lockPeriod"`
	} `json:"staking"`
	Refuel struct {
		NativeMinAmount float64  `json:"nativeMinAmount"`
		NativeMaxAmount float64  `json:"nativeMaxAmount"`
		FromNetwork     string   `json:"fromNetwork"`
		ToNetworks      []string `json:"toNetworks"`
	} `json:"refuel"`
	Rpc struct {
		Bsc      string `json:"bsc"`
		Avax     string `json:"avax"`
		Polygon  string `json:"polygon"`
		Optimism string `json:"optimism"`
		Fantom   string `json:"fantom"`
		Arbitrum string `json:"arbitrum"`
		Aptos    string `json:"aptos"`
		CoreDao  string `json:"coredao"`
	} `json:"rpc"`
}

type BridgeSettings struct {
	FromNetworks        []string `json:"fromNetworks"`
	ToNetworks          []string `json:"toNetworks"`
	BridgeAmountPercent float64  `json:"bridgeAmountPercent"`
}

func ReadSettings(filePath string) Settings {
	configFile, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return Settings{}
	}
	defer configFile.Close()

	var config Settings
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&config); err != nil {
		return Settings{}
	}

	return config
}
