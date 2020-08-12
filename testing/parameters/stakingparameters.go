package parameters

import (
	"strings"

	sdkNetworkTypes "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/types/network"
)

// StakingParameters - represents the test case staking tx parameters
type StakingParameters struct {
	Count int `yaml:"count"`

	Create     CreateValidatorParameters `yaml:"create"`
	Edit       EditValidatorParameters   `yaml:"edit"`
	Delegation DelegationParameters      `yaml:"delegation"`

	Mode                   string `yaml:"mode"`
	ReuseExistingValidator bool   `yaml:"reuse_existing_validator"`

	Gas     sdkNetworkTypes.Gas `yaml:"gas"`
	Nonce   int                 `yaml:"nonce"`
	Timeout int                 `yaml:"timeout"`
}

// Initialize - initializes and converts values for a given test case
func (params *StakingParameters) Initialize() (err error) {

	if len(params.Mode) > 0 {
		params.Mode = strings.ToLower(params.Mode)
	}

	if err = params.Create.Initialize(); err != nil {
		return err
	}

	if err = params.Edit.Initialize(); err != nil {
		return err
	}

	if err = params.Delegation.Initialize(); err != nil {
		return err
	}

	// Initialize gas values
	if err = params.Gas.Initialize(); err != nil {
		return err
	}

	return nil
}
