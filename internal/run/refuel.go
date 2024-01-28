package run

import (
	"filthy/internal/constants"
	"fmt"
	"github.com/TwiN/go-color"
	"strings"
)

func PrintRefuelInfo() {
	info := constants.SETTINGS.Refuel

	println()
	println(color.InRed("Settings refuel"))
	println(color.InCyan("  Native min amount: ") + fmt.Sprintf("%.5f", info.NativeMinAmount))
	println(color.InCyan("  Native max amount: ") + fmt.Sprintf("%.5f", info.NativeMaxAmount))
	println(color.InCyan("  From network: ") + info.FromNetwork)
	println(color.InCyan("  To networks: "), strings.Join(info.ToNetworks, ", "))

}
