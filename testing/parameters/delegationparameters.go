package parameters

import (
	sdkNetworkTypes "github.com/harmony-one/go-lib/network/types/network"
	"github.com/harmony-one/go-sdk/pkg/common"
	"github.com/harmony-one/harmony/numeric"
	"github.com/pkg/errors"
)

// DelegationParameters - the parameters for performing delegation
type DelegationParameters struct {
	// Represents the amount an account will be funded with (if omitted, this will be set to the delegation amount)
	RawAmount string      `yaml:"amount"`
	Amount    numeric.Dec `yaml:"-"`

	Delegate   DelegationInstruction `yaml:"delegate"`
	Undelegate DelegationInstruction `yaml:"undelegate"`
}

// DelegationInstruction - represents a delegation or undelegation instruction
type DelegationInstruction struct {
	RawAmount string              `yaml:"amount"`
	Amount    numeric.Dec         `yaml:"-"`
	Gas       sdkNetworkTypes.Gas `yaml:"gas"`
}

// Initialize - initializes the edit staking parameters
func (delegationParams *DelegationParameters) Initialize() error {
	if delegationParams.RawAmount != "" {
		decAmount, err := common.NewDecFromString(delegationParams.RawAmount)
		if err != nil {
			return errors.Wrapf(err, "DelegationParams: Amount")
		}
		delegationParams.Amount = decAmount
	}

	if err := delegationParams.Delegate.Initialize(); err != nil {
		return err
	}

	if err := delegationParams.Undelegate.Initialize(); err != nil {
		return err
	}

	if delegationParams.RawAmount == "" && delegationParams.Delegate.RawAmount != "" {
		delegationParams.Amount = delegationParams.Delegate.Amount
	}

	return nil
}

// Initialize - initializes the edit staking parameters
func (delegationInstruction *DelegationInstruction) Initialize() error {
	if delegationInstruction.RawAmount != "" {
		decAmount, err := common.NewDecFromString(delegationInstruction.RawAmount)
		if err != nil {
			return errors.Wrapf(err, "DelegationInstruction: Amount")
		}
		delegationInstruction.Amount = decAmount
	}

	// Initialize gas values
	if err := delegationInstruction.Gas.Initialize(); err != nil {
		return err
	}

	return nil
}
