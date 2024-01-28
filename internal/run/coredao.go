package run

import (
	"filthy/internal/constants"
	"fmt"
	"github.com/TwiN/go-color"
)

func PrintCoreInfo(chainFrom string) {
	info := constants.SETTINGS.CoreDao

	println()
	println(color.InRed("Settings core dao bridge"))
	println(color.InCyan("  Bridge amount percent: ") + fmt.Sprintf("%.5f", info.BridgeAmountPercent))
	println(color.InCyan("  From network: ") + chainFrom)

}
