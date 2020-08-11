package testing

import (
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/balances"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"

	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
)

// GenerateAndFundAccount - generates an account and funds it from the core funding account
func GenerateAndFundAccount(testCase *TestCase, accountName string, amount ethCommon.Dec, fundingMultiple int64) (sdkAccounts.Account, error) {
	fundingAccountBalance, err := balances.GetShardBalance(config.Configuration.Funding.Account.Address, testCase.StakingParameters.FromShardID)
	if err != nil {
		return sdkAccounts.Account{}, err
	}

	fundingAmount, err := funding.CalculateFundingAmount(amount, fundingMultiple)
	if err != nil {
		return sdkAccounts.Account{}, err
	}
	logger.FundingLog(fmt.Sprintf("Available funding amount in the funding account %s, address: %s is %f", config.Configuration.Funding.Account.Name, config.Configuration.Funding.Account.Address, fundingAccountBalance), testCase.Verbose)

	logger.AccountLog(fmt.Sprintf("Generating a new account: %s", accountName), testCase.Verbose)
	account, err := accounts.GenerateAccount(accountName)
	logger.AccountLog(fmt.Sprintf("Generated account: %s, address: %s", account.Name, account.Address), testCase.Verbose)
	accountStartingBalance, err := balances.GetShardBalance(account.Address, testCase.StakingParameters.FromShardID)
	if err != nil {
		return sdkAccounts.Account{}, err
	}

	if accountStartingBalance.IsNil() {
		return sdkAccounts.Account{}, fmt.Errorf("Can't fetch starting balance for account %s, address: %s in shard %d", account.Name, account.Address, testCase.StakingParameters.FromShardID)
	}

	gasLimit := -1

	if accountStartingBalance.LT(fundingAmount) {
		funding.PerformFundingTransaction(
			&config.Configuration.Funding.Account,
			testCase.Parameters.FromShardID,
			account.Address,
			testCase.Parameters.ToShardID,
			fundingAmount,
			gasLimit,
			config.Configuration.Funding.Gas.Limit,
			config.Configuration.Funding.Gas.Price,
			config.Configuration.Funding.Timeout,
			config.Configuration.Funding.Retry.Attempts,
		)
		accountStartingBalance, err = balances.GetShardBalance(account.Address, testCase.StakingParameters.FromShardID)
		if err != nil {
			return sdkAccounts.Account{}, err
		}
	}

	account.Balance = accountStartingBalance

	logger.BalanceLog(fmt.Sprintf("Account %s, address: %s has a starting balance of %f in shard %d before the test", account.Name, account.Address, accountStartingBalance, testCase.StakingParameters.FromShardID), testCase.Verbose)

	return account, nil
}
