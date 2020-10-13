package validator

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	restaking "github.com/ethereum/go-ethereum/staking/types/restaking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/pkg/errors"
)

// Validator - represents the validator details
type Validator struct {
	ValidatorAddress string `yaml:"-"`
	OperatorAddress  string `yaml:"-"`
	Account          *accounts.Account
	Details          ValidatorDetails `yaml:"details"`
	Commission       Commission       `yaml:"commission"`
	BLSKeys          []crypto.BLSKey  `yaml:"-"`
	Exists           bool

	RawMaximumTotalDelegation string        `yaml:"maximum_total_delegation"`
	MaximumTotalDelegation    ethCommon.Dec `yaml:"-"`

	RawAmount string        `yaml:"amount"`
	Amount    ethCommon.Dec `yaml:"-"`

	EligibilityStatus string `yaml:"eligibility-status"`
}

// ValidatorDetails - represents the validator details
type ValidatorDetails struct {
	Name            string `yaml:"name"`
	Identity        string `yaml:"identity"`
	Website         string `yaml:"website"`
	SecurityContact string `yaml:"security_contact"`
	Details         string `yaml:"details"`
}

// Commission - represents the validator commission settings
type Commission struct {
	RawRate string        `yaml:"rate"`
	Rate    ethCommon.Dec `yaml:"-"`

	RawMaxRate string        `yaml:"max_rate"`
	MaxRate    ethCommon.Dec `yaml:"-"`

	RawMaxChangeRate string        `yaml:"max_change_rate"`
	MaxChangeRate    ethCommon.Dec `yaml:"-"`
}

// Initialize - initializes and converts values for a given validator
func (validator *Validator) Initialize() error {

	if validator.RawMaximumTotalDelegation != "" {
		decMaximumTotalDelegation, err := common.NewDecFromString(validator.RawMaximumTotalDelegation)
		if err != nil {
			return errors.Wrapf(err, "Validator: MaximumTotalDelegation")
		}
		validator.MaximumTotalDelegation = decMaximumTotalDelegation
	}

	if validator.RawAmount != "" {
		decAmount, err := common.NewDecFromString(validator.RawAmount)
		if err != nil {
			return errors.Wrapf(err, "Validator: Amount")
		}
		validator.Amount = decAmount
	}

	// Initialize commission values
	if err := validator.Commission.Initialize(); err != nil {
		return err
	}

	return nil
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

// ToStakingDescription - convert validator details to a suitable format for staking txs
func (validator *Validator) ToStakingDescription() restaking.Description_ {
	return restaking.Description_{
		Name:            validator.Details.Name,
		Identity:        validator.Details.Identity,
		Website:         validator.Details.Website,
		SecurityContact: validator.Details.SecurityContact,
		Details:         validator.Details.Details,
	}
}

// ToCommissionRates - convert validator commission rates to a suitable format for staking txs
func (validator *Validator) ToCommissionRates() restaking.CommissionRates_ {
	return restaking.CommissionRates_{
		Rate:          validator.Commission.Rate,
		MaxRate:       validator.Commission.MaxRate,
		MaxChangeRate: validator.Commission.MaxChangeRate,
	}
}
