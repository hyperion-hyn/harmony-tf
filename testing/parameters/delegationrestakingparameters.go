package parameters

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	sdkNetworkTypes "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/types/network"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/pkg/errors"
)

// DelegationParameters - the parameters for performing delegation
type DelegationRestakingParameters struct {
	// Represents the amount an account will be funded with (if omitted, this will be set to the delegation amount)
	RawAmount string        `yaml:"amount"`
	Amount    ethCommon.Dec `yaml:"-"`

	Delegate   DelegationRestakingInstruction `yaml:"delegate"`
	Undelegate DelegationRestakingInstruction `yaml:"undelegate"`
}

// DelegationInstruction - represents a delegation or undelegation instruction
type DelegationRestakingInstruction struct {
	Map3Node  map3node.Map3Node   `yaml:"map3Node"`
	RawAmount string              `yaml:"amount"`
	Amount    ethCommon.Dec       `yaml:"-"`
	Gas       sdkNetworkTypes.Gas `yaml:"gas"`
}

// Initialize - initializes the edit staking parameters
func (delegationParams *DelegationRestakingParameters) Initialize() error {
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

	if delegationParams.RawAmount == "" && delegationParams.Delegate.Map3Node.RawAmount != "" {
		delegationParams.Amount = delegationParams.Delegate.Map3Node.Amount
	}

	return nil
}

// Initialize - initializes the edit staking parameters
func (delegationInstruction *DelegationRestakingInstruction) Initialize() error {
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
	if err := delegationInstruction.Map3Node.Initialize(); err != nil {
		return err
	}

	GenerateMap3NodeUniqueDetails(&delegationInstruction.Map3Node.Details)

	return nil
}
