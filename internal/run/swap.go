package run

import (
	"filthy/internal/constants"
	"fmt"
	"github.com/TwiN/go-color"
)

func PrintSwapInfo() {
	info := constants.SETTINGS.Swap

	println()
	println(color.InRed("Settings swap"))
	println(color.InCyan("  Chain: ") + info.Chain)
	println(color.InCyan("  Token From: ") + info.TokenFrom)
	println(color.InCyan("  Token To: ") + info.TokenTo)
	println(color.InCyan("  Percent of token balance: "), fmt.Sprintf("%.2f%%", info.PercentOfUsdc))

}
