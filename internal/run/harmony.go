package run

import (
	"filthy/internal/constants"
	"fmt"
	"github.com/TwiN/go-color"
)

func PrintHarmonyInfo() {
	info := constants.SETTINGS.Harmony

	println()
	println(color.InRed("Settings testnet bridge"))
	println(color.InCyan("  Min amount native token: ") + fmt.Sprintf("%.5f", info.NativeMinAmount))
	println(color.InCyan("  Max amount native token: ") + fmt.Sprintf("%.5f", info.NativeMaxAmount))
	println(color.InCyan("  From network: ") + "bsc")

}
