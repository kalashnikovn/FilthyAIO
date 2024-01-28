package run

import (
	"filthy/internal/constants"
	"fmt"
	"github.com/TwiN/go-color"
)

func PrintTestnetInfo() {
	info := constants.SETTINGS.Testnet

	println()
	println(color.InRed("Settings testnet bridge"))
	println(color.InCyan("  Min amount ETH: ") + fmt.Sprintf("%.5f", info.NativeMinAmount))
	println(color.InCyan("  Max amount ETH: ") + fmt.Sprintf("%.5f", info.NativeMaxAmount))
	println(color.InCyan("  From network: ") + info.FromNetwork)

}
