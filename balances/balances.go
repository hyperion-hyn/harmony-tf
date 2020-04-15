package balances

import (
	"fmt"

	sdkAccounts "github.com/harmony-one/go-lib/accounts"
	"github.com/harmony-one/go-sdk/pkg/common"
	"github.com/harmony-one/harmony-tf/config"
	"github.com/harmony-one/harmony/numeric"
)

// GetShardBalance - gets the balance for a given address and shard
func GetShardBalance(address string, shardID uint32) (numeric.Dec, error) {
	return config.Configuration.Network.API.GetShardBalance(address, shardID)
}

// GetShardBalanceWithRetries - gets the balance for a given address and shard with auto retry upon failure
func GetShardBalanceWithRetries(address string, shardID uint32, attempts int) (balance numeric.Dec, err error) {
	for {
		attempts--

		balance, err = GetShardBalance(address, shardID)
		if err == nil && !balance.IsNil() {
			return balance, nil
		}

		if attempts <= 0 {
			break
		}
	}

	return balance, err
}

// FilterMinimumBalanceAccounts - Filters out a list of accounts without any balance
func FilterMinimumBalanceAccounts(accounts []sdkAccounts.Account, minimumBalance float64) (hasFunds []sdkAccounts.Account, missingFunds []sdkAccounts.Account, err error) {
	for _, account := range accounts {
		totalBalance, err := config.Configuration.Network.API.GetTotalBalance(account.Address)

		if err != nil {
			return nil, nil, err
		}

		decMinimumBalance, err := common.NewDecFromString(fmt.Sprintf("%f", minimumBalance))
		if err != nil {
			return nil, nil, err
		}

		if totalBalance.GT(decMinimumBalance) {
			hasFunds = append(hasFunds, account)
		} else {
			missingFunds = append(missingFunds, account)
		}
	}

	return hasFunds, missingFunds, nil
}

// OutputBalanceStatusForAddresses - outputs balance status
func OutputBalanceStatusForAddresses(accounts []sdkAccounts.Account, minimumBalance float64) {
	hasFunds, missingFunds, err := FilterMinimumBalanceAccounts(accounts, minimumBalance)

	if err == nil {
		fmt.Println(fmt.Sprintf("\nThe following keys hold sufficient funds >%f:", minimumBalance))
		for _, address := range hasFunds {
			fmt.Println(address)
		}

		fmt.Println(fmt.Sprintf("\nThe following keys don't hold sufficient funds of >%f:", minimumBalance))
		for _, address := range missingFunds {
			fmt.Println(address)
		}
	}
}
