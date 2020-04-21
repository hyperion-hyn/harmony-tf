package funding

import (
	"fmt"

	"github.com/harmony-one/harmony-tf/balances"
	"github.com/harmony-one/harmony-tf/config"
	"github.com/harmony-one/harmony/numeric"
	"github.com/pkg/errors"
)

// CalculateFundingDetails - retrieves the funding account balance and calculates the required funding amount
func CalculateFundingDetails(amount numeric.Dec, multiple int64, shardID uint32) (balance numeric.Dec, requiredFunding numeric.Dec, err error) {
	balance, err = RetrieveFundingAccountBalance(shardID)
	if err != nil {
		return numeric.NewDec(0), numeric.NewDec(0), err
	}

	requiredFunding, err = CalculateFundingAmount(amount, multiple)
	if err != nil {
		return numeric.NewDec(0), numeric.NewDec(0), err
	}

	err = VerifyFundingIsPossible(balance, requiredFunding)
	if err != nil {
		return numeric.NewDec(0), numeric.NewDec(0), err
	}

	return balance, requiredFunding, nil
}

// RetrieveFundingAccountBalance - retrieves the balance of the funding account in a specific shard
func RetrieveFundingAccountBalance(shardID uint32) (numeric.Dec, error) {
	fundingAccountBalance, err := balances.GetShardBalance(config.Configuration.Funding.Account.Address, shardID)
	if err != nil {
		err = errors.Wrapf(
			err,
			fmt.Sprintf("Failed to fetch latest account balance for the funding account %s, address: %s",
				config.Configuration.Funding.Account.Name,
				config.Configuration.Funding.Account.Address,
			),
		)

		return numeric.NewDec(0), err
	}

	if fundingAccountBalance.IsNil() || fundingAccountBalance.IsZero() {
		err = errors.Wrapf(
			err,
			fmt.Sprintf("Funding account %s, address: %s doesn't have a sufficient balance in shard %d - balance: %f",
				config.Configuration.Funding.Account.Name,
				config.Configuration.Funding.Account.Address,
				shardID,
				fundingAccountBalance,
			),
		)

		return numeric.NewDec(0), err
	}

	return fundingAccountBalance, nil
}

// CalculateFundingAmount - sets up the initial funding account
func CalculateFundingAmount(amount numeric.Dec, multiple int64) (totalFundingAmount numeric.Dec, err error) {
	if amount.IsNil() || amount.IsNegative() {
		return numeric.NewDec(-1), fmt.Errorf("amount %f can't be nil or negative", amount)
	}

	decCount := numeric.NewDec(multiple)
	totalAmount := decCount.Mul(amount)
	gasCostAmount := decCount.Mul(config.Configuration.Network.Gas.Cost)
	totalFundingAmount = totalAmount.Add(gasCostAmount)

	return totalFundingAmount, nil
}

// VerifyFundingIsPossible - verifies that funding is possible - otherwise return an error
func VerifyFundingIsPossible(balance numeric.Dec, requiredFunding numeric.Dec) error {
	if requiredFunding.GT(balance) {
		return fmt.Errorf(
			"the required funding amount %f is more than what is currently available (%f) in the funding account %s, address: %s ",
			requiredFunding,
			balance,
			config.Configuration.Funding.Account.Name,
			config.Configuration.Funding.Account.Address,
		)
	}

	return nil
}

// InsufficientBalance - checks if a balance is insufficient compared to an expected balance
func InsufficientBalance(balance numeric.Dec, expectedBalance numeric.Dec) bool {
	return (balance.IsNil() || balance.IsZero() || balance.IsNegative() || balance.LT(expectedBalance))
}
