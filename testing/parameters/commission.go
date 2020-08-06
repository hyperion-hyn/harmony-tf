package parameters

import (
	"github.com/harmony-one/harmony/numeric"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/pkg/errors"
)

// Commission - represents the commission parameters
type Commission struct {
	RawRate string      `yaml:"rate"`
	Rate    numeric.Dec `yaml:"-"`

	RawMaxRate string      `yaml:"max_rate"`
	MaxRate    numeric.Dec `yaml:"-"`

	RawMaxChangeRate string      `yaml:"max_change_rate"`
	MaxChangeRate    numeric.Dec `yaml:"-"`
}

// Initialize - initializes and converts values for a given test case
func (commission *Commission) Initialize() error {
	if commission.RawRate != "" {
		decRate, err := common.NewDecFromString(commission.RawRate)
		if err != nil {
			return errors.Wrapf(err, "Commission: Rate")
		}
		commission.Rate = decRate
	}

	if commission.RawMaxRate != "" {
		decMaxRate, err := common.NewDecFromString(commission.RawMaxRate)
		if err != nil {
			return errors.Wrapf(err, "Commission: MaxRate")
		}
		commission.MaxRate = decMaxRate
	}

	if commission.RawMaxChangeRate != "" {
		decMaxChangeRate, err := common.NewDecFromString(commission.RawMaxChangeRate)
		if err != nil {
			return errors.Wrapf(err, "Commission: MaxChangeRate")
		}
		commission.MaxChangeRate = decMaxChangeRate
	}

	return nil
}
