package funding

import (
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"

	"github.com/hyperion-hyn/hyperion-tf/balances"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/pkg/errors"
)

// CalculateFundingDetails - retrieves the funding account balance and calculates the required funding amount
func CalculateFundingDetails(amount ethCommon.Dec, multiple int64) (balance ethCommon.Dec, requiredFunding ethCommon.Dec, err error) {
	balance, err = RetrieveFundingAccountBalance()
	if err != nil {
		return ethCommon.NewDec(0), ethCommon.NewDec(0), err
	}

	requiredFunding, err = CalculateFundingAmount(amount, multiple)
	if err != nil {
		return ethCommon.NewDec(0), ethCommon.NewDec(0), err
	}

	err = VerifyFundingIsPossible(balance, requiredFunding)
	if err != nil {
		return ethCommon.NewDec(0), ethCommon.NewDec(0), err
	}

	return balance, requiredFunding, nil
}

// RetrieveFundingAccountBalance - retrieves the balance of the funding account in a specific shard
func RetrieveFundingAccountBalance() (ethCommon.Dec, error) {
	fundingAccountBalance, err := balances.GetBalance(config.Configuration.Funding.Account.Address)
	if err != nil {
		err = errors.Wrapf(
			err,
			fmt.Sprintf("Failed to fetch latest account balance for the funding account %s, address: %s",
				config.Configuration.Funding.Account.Name,
				config.Configuration.Funding.Account.Address,
			),
		)

		return ethCommon.NewDec(0), err
	}

	if fundingAccountBalance.IsNil() || fundingAccountBalance.IsZero() {
		err = errors.Wrapf(
			err,
			fmt.Sprintf("Funding account %s, address: %s doesn't have a sufficient balance - balance: %f",
				config.Configuration.Funding.Account.Name,
				config.Configuration.Funding.Account.Address,
				fundingAccountBalance,
			),
		)

		return ethCommon.NewDec(0), err
	}

	return fundingAccountBalance, nil
}

// CalculateFundingAmount - sets up the initial funding account
func CalculateFundingAmount(amount ethCommon.Dec, multiple int64) (totalFundingAmount ethCommon.Dec, err error) {
	if amount.IsNil() || amount.IsNegative() {
		return ethCommon.NewDec(-1), fmt.Errorf("amount %f can't be nil or negative", amount)
	}

	decCount := ethCommon.NewDec(multiple)
	totalAmount := decCount.Mul(amount)
	gasCostAmount := decCount.Mul(config.Configuration.Network.Gas.Cost)
	totalFundingAmount = totalAmount.Add(gasCostAmount)

	return totalFundingAmount, nil
}

// VerifyFundingIsPossible - verifies that funding is possible - otherwise return an error
func VerifyFundingIsPossible(balance ethCommon.Dec, requiredFunding ethCommon.Dec) error {
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
func InsufficientBalance(balance ethCommon.Dec, expectedBalance ethCommon.Dec) bool {
	return (balance.IsNil() || balance.IsZero() || balance.IsNegative() || balance.LT(expectedBalance))
}
