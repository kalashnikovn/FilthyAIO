package account

import (
	"context"
	"filthy/internal/constants"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"math/rand"
	"time"
)

type Helper struct{}

func (acc Helper) FindChainWithMoney(token string, chains []string, account common.Address) (MoneyChain, error) {
	balances, err := acc.GetTokenBalances(token, chains, account)
	if err != nil {
		return MoneyChain{}, err
	}

	balance := GetMaxBalance(balances)
	return balance, nil
}

func (acc Helper) GetTokenBalances(token string, chains []string, account common.Address) (map[string]Balance, error) {
	balances := make(map[string]Balance)

	for _, chain := range chains {
		btcAddress := common.HexToAddress(constants.CONTRACTS[token][chain])
		client := constants.CLIENTS[chain]

		amount, err := acc.balanceOf(client, btcAddress, account)
		if err != nil {
			constants.Logger.Error("Поменяй ноду: ", chain, " Скорее всего ошибка при подключении к ней...")
			return map[string]Balance{}, err
		}

		decimal := constants.TOKEN_DECIMALS[token][chain]

		balances[chain] = Balance{
			prettyAmount(amount, decimal),
			amount,
		}
	}

	return balances, nil
}

func (acc Helper) balanceOf(client *ethclient.Client, contractAddress common.Address, account common.Address) (*big.Int, error) {
	usdcAbi, err := ReadAbi(constants.USDC_ABI)
	if err != nil {
		return nil, err
	}
	callData, err := usdcAbi.Pack("balanceOf", account)
	if err != nil {
		return nil, err
	}

	res, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, err
	}

	result, err := usdcAbi.Unpack("balanceOf", res)
	if err != nil {
		return nil, err
	}

	return result[0].(*big.Int), nil

}

func (acc Helper) Approve(chainFrom string, amountToApprove *big.Int, token, bridge string, account Wallet) (*types.Receipt, error) {
	usdcContract := common.HexToAddress(constants.CONTRACTS[token][chainFrom])
	stargateContract := common.HexToAddress(constants.BRIDGE_CONTRACTS[bridge][chainFrom])

	//spew.Dump(stargateContract)

	client := constants.CLIENTS[chainFrom]

	nonce, err := client.PendingNonceAt(context.Background(), account.PublicKey)
	if err != nil {
		return nil, err
	}

	contractAbi, err := ReadAbi(constants.USDC_ABI)
	if err != nil {
		return nil, err
	}

	amount := new(big.Int).Mul(amountToApprove, big.NewInt(10))

	encodedData, err := contractAbi.Pack("approve", stargateContract, amount)
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:    &usdcContract,
		From:  account.PublicKey,
		Value: big.NewInt(0),
		Data:  encodedData,
	}

	//gasPrice, err := client.SuggestGasPrice(context.Background())
	gasPrice, err := GetGasPrice(client, chainFrom)
	if err != nil {
		return nil, err
	}

	gasLimit, err := GetGasLimit(client, msg)
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(
		nonce,
		usdcContract,
		big.NewInt(0),
		gasLimit,
		gasPrice,
		encodedData,
	)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), account.PrivateKey)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %v\n", err)
		return nil, err
	}

	//fmt.Println("Approve", signedTx.Hash())

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	hash, err := WaitMined(client, signedTx)
	if err != nil {
		return nil, err
	}

	return hash, nil

}

func (acc Helper) CalculateMinAmount(n *big.Int, percent float64) *big.Int {
	percentage := new(big.Rat).SetFloat64(percent / 100)

	m := new(big.Int).Mul(n, percentage.Num())
	quotient := m.Div(m, percentage.Denom())

	rounded, _ := quotient.DivMod(quotient, big.NewInt(1), new(big.Int))

	if rounded.Cmp(n) == 1 {
		return n
	}

	return rounded
}

func (acc Helper) GetAmounts(balance Balance, bridge string) SwapArgs {
	percent := constants.BRIDGE_AMOUNT_PERCENTS[bridge]
	slippage := constants.SETTINGS.Accounts.Slippage

	amount := acc.CalculateMinAmount(balance.BigAmount, percent)

	minAmount := acc.CalculateMinAmount(amount, 100.0-slippage)

	return SwapArgs{
		Amount:    amount,
		MinAmount: minAmount,
	}

}

func (acc Helper) Allowance(chainFrom, token, bridge string, account Wallet) (*big.Int, error) {
	client := constants.CLIENTS[chainFrom]
	usdcAbi, err := ReadAbi(constants.USDC_ABI)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(constants.CONTRACTS[token][chainFrom])
	stargateContract := common.HexToAddress(constants.BRIDGE_CONTRACTS[bridge][chainFrom])

	callData, err := usdcAbi.Pack("allowance", account.PublicKey, stargateContract)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	res, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddress,
		Data: callData,
	}, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	result, err := usdcAbi.Unpack("allowance", res)
	if err != nil {
		return nil, err
	}

	return result[0].(*big.Int), nil
}

func (acc Helper) GetRandomActivityDelay() int {
	activity := constants.SETTINGS.Accounts.DelayActivity
	min := activity[0]
	max := activity[1]
	rand.Seed(time.Now().UnixNano())
	intn := rand.Intn(max-min+1) + min
	return intn
}
