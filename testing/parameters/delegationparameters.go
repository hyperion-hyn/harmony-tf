package parameters

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	sdkNetworkTypes "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/types/network"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/pkg/errors"
)

// DelegationParameters - the parameters for performing delegation
type DelegationParameters struct {
	// Represents the amount an account will be funded with (if omitted, this will be set to the delegation amount)
	RawAmount string        `yaml:"amount"`
	Amount    ethCommon.Dec `yaml:"-"`

	Delegate   DelegationInstruction `yaml:"delegate"`
	Undelegate DelegationInstruction `yaml:"undelegate"`
	Terminate  DelegationInstruction `yaml:"terminate"`
	Renew      RenewInstruction      `yaml:"renew"`
}

// DelegationInstruction - represents a delegation or undelegation instruction
type DelegationInstruction struct {
	RawAmount string              `yaml:"amount"`
	Amount    ethCommon.Dec       `yaml:"-"`
	Gas       sdkNetworkTypes.Gas `yaml:"gas"`
	WaitEpoch int64               `yaml:"wait_epoch"`
}

type RenewInstruction struct {
	OperatorWaitEpoch        int64               `yaml:"operator_wait_epoch"`
	OperatorRenew            bool                `yaml:"operator_renew"`
	OperatorSendRenew        bool                `yaml:"operator_send_renew"`
	OperatorRawCommission    string              `yaml:"operator_commission"`
	OperatorCommission       *ethCommon.Dec      `yaml:"-"`
	ParticipantWaitEpoch     int64               `yaml:"participant_wait_epoch"`
	ParticipantRenew         bool                `yaml:"participant_renew"`
	ParticipantSendRenew     bool                `yaml:"participant_send_renew"`
	ParticipantRawCommission string              `yaml:"participant_commission"`
	ParticipantCommission    *ethCommon.Dec      `yaml:"-"`
	Gas                      sdkNetworkTypes.Gas `yaml:"gas"`
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

	if err := delegationParams.Renew.Initialize(); err != nil {
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

// Initialize - initializes the edit staking parameters
func (renewInstruction *RenewInstruction) Initialize() error {
	if renewInstruction.OperatorRawCommission != "" {
		decCommission, err := common.NewDecFromString(renewInstruction.OperatorRawCommission)
		if err != nil {
			return errors.Wrapf(err, "Map3Node: Commission")
		}
		renewInstruction.OperatorCommission = &decCommission
	}
	if renewInstruction.ParticipantRawCommission != "" {
		decCommission, err := common.NewDecFromString(renewInstruction.ParticipantRawCommission)
		if err != nil {
			return errors.Wrapf(err, "Map3Node: Commission")
		}
		renewInstruction.ParticipantCommission = &decCommission
	}

	// Initialize gas values
	if err := renewInstruction.Gas.Initialize(); err != nil {
		return err
	}
	return nil
}
