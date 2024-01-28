package account

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"sort"
	"strings"
	"time"
)

type Balance struct {
	UiAmount  float64
	BigAmount *big.Int
}

type MoneyChain struct {
	Chain string
	Balance
}

type SwapArgs struct {
	Amount, MinAmount *big.Int
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

func ReadAbi(abiString string) (abi.ABI, error) {
	myContractAbi, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		return abi.ABI{}, err
	}

	return myContractAbi, nil
}

func prettyAmount(a *big.Int, b uint) float64 {
	base := big.NewInt(10)
	power := big.NewInt(int64(b))
	power.Exp(base, power, nil)

	fa := new(big.Float).SetInt(a)
	fp := new(big.Float).SetInt(power)

	fa.Quo(fa, fp)

	f, _ := fa.Float64()

	return f
}

func СonvertAddress(address string) string {
	address = strings.TrimPrefix(address, "0x")

	var result [32]byte

	addressBytes := common.Hex2Bytes(address)

	copy(result[32-len(addressBytes):], addressBytes)

	// возвращаем адрес в нужном формате
	return "0x" + common.Bytes2Hex(result[:])
}

func GetMaxBalance(balances map[string]Balance) MoneyChain {
	var balanceSlice []Balance
	var balanceKeys []string
	for key, balance := range balances {
		balanceSlice = append(balanceSlice, balance)
		balanceKeys = append(balanceKeys, key)
	}

	// Определяем функцию сортировки
	sort.Slice(balanceSlice, func(i, j int) bool {
		return balanceSlice[i].UiAmount > balanceSlice[j].UiAmount
	})

	// Получаем структуру с наибольшим значением и ее ключ в мапе
	maxBalance := balanceSlice[0]
	maxKey := ""
	for key, balance := range balances {
		if balance == maxBalance {
			maxKey = key
			break
		}
	}

	return MoneyChain{
		maxKey,
		maxBalance,
	}
}
