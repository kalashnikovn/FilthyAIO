package run

import (
	"filthy/internal/constants"
	"fmt"
	"github.com/TwiN/go-color"
	"strings"
)

func PrintAptosInfo() {
	info := constants.SETTINGS.Aptos

	println()
	println(color.InRed("Settings to Aptos bridge"))
	println(color.InCyan("  From Networks: ") + strings.Join(info.FromNetworks, ", "))
	println(color.InCyan("  Bridge Percent Amount of USDC: "), fmt.Sprintf("%.3f%%", info.BridgeAmountPercent))

}

func PrintFromAptosInfo() {
	println()
	println(color.InRed("Settings from Aptos bridge"))
	println(color.InCyan("  Информация: модуль выведет все найденные USDC с аптос аккаунтов в сеть Avalanche"))
}
