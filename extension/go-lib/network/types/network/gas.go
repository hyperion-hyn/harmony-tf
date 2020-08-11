package network

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/pkg/errors"
)

// Gas - represents the gas settings
type Gas struct {
	RawCost  string        `json:"cost" yaml:"cost"`
	Cost     ethCommon.Dec `json:"-" yaml:"-"`
	Limit    int64         `json:"limit" yaml:"limit"`
	RawPrice string        `json:"price" yaml:"price"`
	Price    ethCommon.Dec `json:"-" yaml:"-"`
}

// Initialize - convert the raw values to their appropriate ethCommon.Dec values
func (gas *Gas) Initialize() error {
	if gas.RawCost != "" {
		decCost, err := common.NewDecFromString(gas.RawCost)
		if err != nil {
			return errors.Wrapf(err, "Gas: Cost")
		}
		gas.Cost = decCost
	}

	if gas.RawPrice != "" {
		decPrice, err := common.NewDecFromString(gas.RawPrice)
		if err != nil {
			return errors.Wrapf(err, "Gas: Price")
		}
		gas.Price = decPrice
	}

	if gas.Limit == 0 {
		gas.Limit = -1
	}

	if gas.Price.IsNil() || gas.Price.IsZero() || gas.Price.IsNegative() {
		gas.Price = ethCommon.NewDec(1)
	}

	return nil
}
