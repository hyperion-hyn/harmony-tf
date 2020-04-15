package funding

import (
	"fmt"

	"github.com/harmony-one/harmony-tf/config"
	"github.com/harmony-one/harmony/numeric"
)

// CalculateFundingAmount - sets up the initial funding account
func CalculateFundingAmount(amount numeric.Dec, balance numeric.Dec, count int64) (totalFundingAmount numeric.Dec, err error) {
	if amount.IsNil() || amount.IsNegative() {
		return numeric.NewDec(-1), fmt.Errorf("amount %f can't be nil or negative", amount)
	}

	decCount := numeric.NewDec(count)
	totalAmount := decCount.Mul(amount)
	gasCostAmount := decCount.Mul(config.Configuration.Network.Gas.Cost)
	totalFundingAmount = totalAmount.Add(gasCostAmount)

	/*if totalFundingAmount.GT(balance) {
		return numeric.NewDec(-1), fmt.Errorf(
			"the required funding amount %f is more than what is currently available (%f) in the account %s, address: %s ",
			totalFundingAmount,
			balance,
			config.Configuration.Funding.Account.Name,
			config.Configuration.Funding.Account.Address,
		)
	}*/

	return totalFundingAmount, nil
}
