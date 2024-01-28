package account

import (
	"context"
	"filthy/internal/constants"
	"filthy/internal/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

func ValidateFee(chainFrom string, gasPrice *big.Int, gasLimit uint64, value *big.Int) bool {
	userFee := constants.SETTINGS.Fee[chainFrom]
	userLimit := ToWei(userFee)
	totalStargateFee := CalculateFee(gasPrice, gasLimit, value)

	cmp := userLimit.Cmp(totalStargateFee)
	if cmp == 1 {
		return true
	}

	return false
}

func CalculateFee(gasPrice *big.Int, gasLimit uint64, value *big.Int) *big.Int {
	swapFeeGwei := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
	totalFee := big.NewInt(0).Add(swapFeeGwei, value)

	return totalFee
}

func GetGasLimit(client *ethclient.Client, msg ethereum.CallMsg) (uint64, error) {
	gasLim, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, err
	}

	increased := utils.IncreaseByPercent(gasLim, 1.01)

	return increased, nil
}

func ToWei(amount float64) *big.Int {
	weiAmount := big.NewFloat(amount)
	weiAmount = new(big.Float).Mul(weiAmount, big.NewFloat(1e18))

	wei, _ := weiAmount.Int(nil)
	return wei
}

func GetGasPrice(client *ethclient.Client, chainFrom string) (*big.Int, error) {
	if chainFrom == "bsc" {
		bscGwei := constants.SETTINGS.Accounts.BscGwei
		return gweiToWei(bscGwei), nil
	} else {
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, err
		}
		return gasPrice, nil
	}

}

func gweiToWei(gwei int64) *big.Int {
	wei := big.NewInt(gwei)
	wei.Mul(wei, big.NewInt(1_000_000_000))
	return wei
}
