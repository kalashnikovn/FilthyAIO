package run

import (
	"filthy/internal/constants"
	"fmt"
	"github.com/TwiN/go-color"
	"strings"
)

func PrintRandomInfo(bridges []string) {
	println()
	println(color.InRed("Settings bridge"))
	for _, bridge := range bridges {
		if bridge == "testnet" {
			info := constants.SETTINGS.Testnet
			println("  " + color.InBlue(strings.Title(bridge)))
			println(color.InCyan("    Min amount ETH: ") + fmt.Sprintf("%.5f", info.NativeMinAmount))
			println(color.InCyan("    Max amount ETH: ") + fmt.Sprintf("%.5f", info.NativeMaxAmount))
			println(color.InCyan("    From network: ") + info.FromNetwork)
		} else if bridge == "harmony" {
			info := constants.SETTINGS.Harmony

			println("  " + color.InBlue(strings.Title(bridge)))
			println(color.InCyan("    Min amount native token: ") + fmt.Sprintf("%.5f", info.NativeMinAmount))
			println(color.InCyan("    Max amount native token: ") + fmt.Sprintf("%.5f", info.NativeMaxAmount))
			println(color.InCyan("    From network: ") + "bsc")
		} else {
			info := constants.BRIDGE_SETTINGS[bridge]
			println("  " + color.InBlue(strings.Title(bridge)))
			println(color.InCyan("    From Networks: ") + strings.Join(info.FromNetworks, ", "))
			println(color.InCyan("    To Networks: ") + strings.Join(info.ToNetworks, ", "))
			println(color.InCyan("    Bridge Percent Amount: "), fmt.Sprintf("%.3f%%", info.BridgeAmountPercent))
		}
	}

}
