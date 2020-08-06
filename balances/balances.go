package balances

import (
	"fmt"
	"time"

	"github.com/harmony-one/harmony/numeric"
	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
)

// GetShardBalance - gets the balance for a given address and shard
func GetShardBalance(address string, shardID uint32) (numeric.Dec, error) {
	return config.Configuration.Network.API.GetShardBalance(address, shardID)
}

// GetNonZeroShardBalance - gets the balance for a given address and shard with auto retry upon failure/balance being nil/balance being zero
func GetNonZeroShardBalance(address string, shardID uint32) (balance numeric.Dec, err error) {
	attempts := config.Configuration.Network.Balances.Retry.Attempts

	for {
		attempts--

		balance, err = GetShardBalance(address, shardID)
		if err == nil && !balance.IsNil() && !balance.IsZero() {
			return balance, nil
		}

		if attempts <= 0 {
			break
		}

		time.Sleep(time.Duration(config.Configuration.Network.Balances.Retry.Wait) * time.Second)
	}

	return balance, err
}

// GetExpectedShardBalance - retries to fetch the balance for a given address in a given shard until an expected balance exists
func GetExpectedShardBalance(address string, shardID uint32, expectedBalance numeric.Dec) (balance numeric.Dec, err error) {
	attempts := config.Configuration.Network.Balances.Retry.Attempts

	for {
		attempts--

		balance, err = GetShardBalance(address, shardID)
		if err == nil && !balance.IsNil() && !balance.IsZero() && balance.GTE(expectedBalance) {
			return balance, nil
		}

		if attempts <= 0 {
			return numeric.NewDec(0), fmt.Errorf("failed to retrieve expected balance %f for address %s in shard %d", expectedBalance, address, shardID)
		}

		time.Sleep(time.Duration(config.Configuration.Network.Balances.Retry.Wait) * time.Second)
	}
}

// FilterMinimumBalanceAccounts - Filters out a list of accounts without any balance
func FilterMinimumBalanceAccounts(accounts []sdkAccounts.Account, minimumBalance numeric.Dec) (hasFunds []sdkAccounts.Account, missingFunds []sdkAccounts.Account, err error) {
	for _, account := range accounts {
		totalBalance, err := config.Configuration.Network.API.GetTotalBalance(account.Address)

		if err != nil {
			return nil, nil, err
		}

		if totalBalance.GT(minimumBalance) {
			hasFunds = append(hasFunds, account)
		} else {
			missingFunds = append(missingFunds, account)
		}
	}

	return hasFunds, missingFunds, nil
}

// OutputBalanceStatusForAddresses - outputs balance status
func OutputBalanceStatusForAddresses(accounts []sdkAccounts.Account, minimumBalance numeric.Dec) {
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
