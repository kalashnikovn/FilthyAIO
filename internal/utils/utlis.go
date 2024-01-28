package utils

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"filthy/internal/constants"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

type Pair struct {
	Key   string
	Value float64
}

func SortByValue(myMap map[string]float64) []Pair {
	pairs := make([]Pair, len(myMap))
	i := 0
	for key, value := range myMap {
		pairs[i] = Pair{key, value}
		i++
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value
	})
	return pairs
}

func RandomChain(chains []string, exclude string) (string, error) {
	if len(chains) == 1 && chains[0] == exclude {
		return "", errors.New("chainFrom = chainTo, такой бридж невозможен")
	}

	rand.Seed(time.Now().UnixNano())

	for {
		randIndex := rand.Intn(len(chains))
		if chains[randIndex] != exclude {
			return chains[randIndex], nil
		}
	}
}

func DecreaseByPercent(number, percent float64) float64 {
	return number * (percent / 100)
}

func ToBigFloat(num float64, exp uint) *big.Float {
	var f big.Float
	f.SetPrec(256)
	f.SetMode(big.ToNearestEven)
	f.SetInt(big.NewInt(int64(num * math.Pow10(int(exp)))))
	return &f
}

func IncreaseByPercent(num uint64, percent float64) uint64 {
	return uint64(float64(num) * percent)
}

func PrivateKeyToString(privateKey *ecdsa.PrivateKey) string {
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)

	return privateKeyHex
}

func AppendToFile(filename, text string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, text)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func ReadAbi(abiString string) (abi.ABI, error) {
	myContractAbi, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		return abi.ABI{}, err
	}

	return myContractAbi, nil
}

func ToWei(iamount interface{}, decimals int) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}

func ValidateFee(chainFrom string, gasPrice *big.Int, gasLimit uint64, value *big.Int) bool {
	userFee := constants.SETTINGS.Fee[chainFrom]
	userLimit := ToWei(userFee, 18)
	totalStargateFee := CalculateFee(gasPrice, gasLimit, value)

	//spew.Dump(totalStargateFee)

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

	increased := IncreaseByPercent(gasLim, 1.01)

	return increased, nil
}

func WaitMined(client *ethclient.Client, signedTx *types.Transaction) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	mined, err := bind.WaitMined(ctx, client, signedTx)
	if err != nil {
		return nil, err
	}

	return mined, nil
}

func MultiplyBigInt(value *big.Int, multiplier float64) (*big.Int, error) {
	// Преобразование big.Int в decimal.Decimal
	decimalValue := decimal.NewFromBigInt(value, 0)

	// Преобразование множителя в decimal.Decimal
	decimalMultiplier := decimal.NewFromFloat(multiplier)

	// Умножение значения на множитель
	result := decimalValue.Mul(decimalMultiplier)

	// Преобразование обратно в big.Int
	resultBigInt := result.BigInt()

	return resultBigInt, nil
}

func GetRandomFloat(min, max float64) float64 {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Float64()*(max-min)
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
