package main

import (
	"filthy/internal/account"
	"filthy/internal/aptos"
	"filthy/internal/client"
	"filthy/internal/constants"
	"filthy/internal/run"
	"fmt"
	"github.com/gorilla/websocket"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	clearTerminal()
	fmt.Println(color.InRed(constants.LOGO))
	println()

	wallets, err := account.NewWallets()
	if err != nil {
		return
	}

	conn, tgName, err := auth.Authorize()
	if err != nil {
		println(color.InRed("Авторизация не удалась...."))
		fmt.Scanln()
		return
	}
	conn.Close()

	println(color.InGreen("Авторизация прошла успешно. Добро пожаловать,"), tgName)

	go readLoop(conn)

	start(wallets)

	fmt.Scanln()

}

func start(wallets []account.Wallet) {
	module := ""
	prompt := &survey.Select{
		Message: "Выбери модуль:",
		Options: []string{"Random", "Swap", "Refuel", "Testnet", "Harmony", "Bridge to Core", "Bridge From Core", "Bridge to Aptos", "Bridge from Aptos", "Staking", "Exit"},
	}
	survey.AskOne(prompt, &module)

	if module == "Random" {
		run.StartRandom(wallets)
	} else if module == "Swap" {
		run.StartSwap(wallets)
	} else if module == "Refuel" {
		run.StartRefuel(wallets)
	} else if module == "Testnet" {
		run.StartTestnet(wallets)
	} else if module == "Bridge to Aptos" {
		aptosClient := client.NewAptos(constants.SETTINGS.Rpc.Aptos)
		aptWallets, err := aptos.NewWallets(aptosClient)
		if err != nil {
			fmt.Println(err)
			return
		}
		run.StartBridgeToAptos(aptWallets)
	} else if module == "Bridge from Aptos" {
		aptosClient := client.NewAptos(constants.SETTINGS.Rpc.Aptos)
		aptWallets, err := aptos.NewWallets(aptosClient)
		if err != nil {
			fmt.Println(err)
			return
		}
		run.StartBridgeFromAptos(aptWallets)
	} else if module == "Harmony" {
		run.StartHarmony(wallets)
	} else if module == "Bridge From Core" {
		run.StartFromCoreDao(wallets)
	} else if module == "Bridge to Core" {
		run.StartCoreDao(wallets)
	} else if module == "Staking" {
		run.StartStaking(wallets)
	} else if module == "Exit" {
		os.Exit(1)
	}

	start(wallets)

}

func readLoop(conn *websocket.Conn) {
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			constants.Logger.Error("Connection closed:", err)
			os.Exit(1)
		}
	}
}

func clearTerminal() {
	osName := runtime.GOOS

	var cmd *exec.Cmd
	if osName == "windows" {
		cmd = exec.Command("fil", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}
