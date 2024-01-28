package run

import (
	"filthy/internal/constants"
	"github.com/TwiN/go-color"
)

func PrintStakingInfo() {
	info := constants.SETTINGS.Staking

	println()
	println(color.InRed("Settings staking"))
	println(color.InCyan("  Chain: ") + info.FromNetwork)

}
