package parameters

import (
	"fmt"
	"github.com/ethereum/go-ethereum/staking/types/restaking"
	"strings"

	sdkNetworkTypes "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/types/network"
	sdkValidator "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking/validator"
	"github.com/hyperion-hyn/hyperion-tf/logger"
)

// EditValidatorParameters - the parameters for editing a validator
type EditValidatorParameters struct {
	Mode      string                 `yaml:"mode"`
	Repeat    uint32                 `yaml:"repeat"`
	Validator sdkValidator.Validator `yaml:"validator"`

	Gas sdkNetworkTypes.Gas `yaml:"gas"`

	Changes EditValidatorChanges `yaml:"-"`

	RandomizeUniqueFields bool `yaml:"randomize_unique_fields"`
}

// EditValidatorChanges - keeps track of what fields have changed
type EditValidatorChanges struct {
	ValidatorName            bool   `yaml:"-"`
	ValidatorIdentity        bool   `yaml:"-"`
	ValidatorWebsite         bool   `yaml:"-"`
	ValidatorSecurityContact bool   `yaml:"-"`
	ValidatorDetails         bool   `yaml:"-"`
	CommissionRate           bool   `yaml:"-"`
	MinimumSelfDelegation    bool   `yaml:"-"`
	MaximumTotalDelegation   bool   `yaml:"-"`
	EligibilityStatus        bool   `yaml:"-"`
	TotalChanged             uint32 `yaml:"-"`
}

// Initialize - initializes the edit staking parameters
func (editParams *EditValidatorParameters) Initialize() error {
	if editParams.Repeat == 0 {
		editParams.Repeat = 1
	}

	if err := editParams.Validator.Initialize(); err != nil {
		return err
	}

	if editParams.RandomizeUniqueFields {
		GenerateUniqueDetails(&editParams.Validator.Details)
	}

	if err := editParams.Gas.Initialize(); err != nil {
		return err
	}

	return nil
}

// DetectChanges - detects which fields have been changed during an edit validator procedure
func (editParams *EditValidatorParameters) DetectChanges(verbose bool) {
	if editParams.Mode != "" {
		editParams.Mode = strings.ToLower(editParams.Mode)
	}

	if editParams.Validator.Details.Name != "" {
		editParams.Changes.ValidatorName = true
		editParams.Changes.TotalChanged++
		logger.StakingLog(fmt.Sprintf("Will update the name of the validator to %s", editParams.Validator.Details.Name), verbose)
	}

	if editParams.Validator.Details.Identity != "" {
		logger.StakingLog(fmt.Sprintf("Will update the identity of the validator to %s", editParams.Validator.Details.Identity), verbose)
		editParams.Changes.ValidatorIdentity = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Validator.Details.Website != "" {
		logger.StakingLog(fmt.Sprintf("Will update the website of the validator to %s", editParams.Validator.Details.Website), verbose)
		editParams.Changes.ValidatorWebsite = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Validator.Details.SecurityContact != "" {
		logger.StakingLog(fmt.Sprintf("Will update the security contact of the validator to %s", editParams.Validator.Details.SecurityContact), verbose)
		editParams.Changes.ValidatorSecurityContact = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Validator.Details.Details != "" {
		logger.StakingLog(fmt.Sprintf("Will update the details of the validator to %s", editParams.Validator.Details.Details), verbose)
		editParams.Changes.ValidatorDetails = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Validator.Commission.RawRate != "" {
		logger.StakingLog(fmt.Sprintf("Will update the commission rate of the validator to %f", editParams.Validator.Commission.Rate), verbose)
		editParams.Changes.CommissionRate = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Validator.RawMinimumSelfDelegation != "" {
		logger.StakingLog(fmt.Sprintf("Will update the minimum self delegation of the validator to %f", editParams.Validator.MinimumSelfDelegation), verbose)
		editParams.Changes.MinimumSelfDelegation = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Validator.RawMaximumTotalDelegation != "" {
		logger.StakingLog(fmt.Sprintf("Will update the maximum total delegation of the validator to %f", editParams.Validator.MaximumTotalDelegation), verbose)
		editParams.Changes.MaximumTotalDelegation = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Validator.EligibilityStatus != "" {
		logger.StakingLog(fmt.Sprintf("Will update the eligibility status of the validator to %s", editParams.Validator.EligibilityStatus), verbose)
		editParams.Changes.EligibilityStatus = true
		editParams.Changes.TotalChanged++
	}
}

// EvaluateChanges - evaluates which changes have taken place and if they were successful
func (editParams *EditValidatorParameters) EvaluateChanges(validatorInfo restaking.ValidatorWrapperRPC, verbose bool) bool {
	successfulChangeCount := uint32(0)

	if editParams.Changes.ValidatorName {
		if validatorInfo.Validator.Description.Name == editParams.Validator.Details.Name {
			logger.StakingLog(fmt.Sprintf("Successfully updated the name of the validator to %s", editParams.Validator.Details.Name), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the name of the validator to %s - returned name is %s", editParams.Validator.Details.Name, validatorInfo.Validator.Description.Name), verbose)
		}
	}

	if editParams.Changes.ValidatorIdentity {
		if validatorInfo.Validator.Description.Identity == editParams.Validator.Details.Identity {
			logger.StakingLog(fmt.Sprintf("Successfully updated the identity of the validator to %s", editParams.Validator.Details.Identity), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the identity of the validator to %s - returned identity is %s", editParams.Validator.Details.Identity, validatorInfo.Validator.Description.Identity), verbose)
		}
	}

	if editParams.Changes.ValidatorWebsite {
		if validatorInfo.Validator.Description.Website == editParams.Validator.Details.Website {
			logger.StakingLog(fmt.Sprintf("Successfully updated the website of the validator to %s", editParams.Validator.Details.Website), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the website of the validator to %s - returned website is %s", editParams.Validator.Details.Website, validatorInfo.Validator.Description.Website), verbose)
		}
	}

	if editParams.Changes.ValidatorSecurityContact {
		if validatorInfo.Validator.Description.SecurityContact == editParams.Validator.Details.SecurityContact {
			logger.StakingLog(fmt.Sprintf("Successfully updated the security contact of the validator to %s", editParams.Validator.Details.SecurityContact), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the security contact of the validator to %s - returned security contact is %s", editParams.Validator.Details.SecurityContact, validatorInfo.Validator.Description.SecurityContact), verbose)
		}
	}

	if editParams.Changes.ValidatorDetails {
		if validatorInfo.Validator.Description.Details == editParams.Validator.Details.Details {
			logger.StakingLog(fmt.Sprintf("Successfully updated the details of the validator to %s", editParams.Validator.Details.Details), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the details of the validator to %s - returned details is %s", editParams.Validator.Details.Details, validatorInfo.Validator.Description.Details), verbose)
		}
	}

	if editParams.Changes.CommissionRate {
		if !validatorInfo.Validator.Commission.CommissionRates.Rate.IsNil() && validatorInfo.Validator.Commission.CommissionRates.Rate.Equal(editParams.Validator.Commission.Rate) {
			logger.StakingLog(fmt.Sprintf("Successfully updated the commission rate of the validator to %f", editParams.Validator.Commission.Rate), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the commission rate of the validator to %f - returned commission rate is %f", editParams.Validator.Commission.Rate, validatorInfo.Validator.Commission.CommissionRates.Rate), verbose)
		}
	}

	if editParams.Changes.MaximumTotalDelegation {
		if validatorInfo.Validator.MaxTotalDelegation != nil && validatorInfo.Validator.MaxTotalDelegation.Cmp(editParams.Validator.MaximumTotalDelegation.BigInt()) == 0 {
			logger.StakingLog(fmt.Sprintf("Successfully updated the maximum total delegation of the validator to %f", editParams.Validator.MaximumTotalDelegation), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the maximum total delegation of the validator to %f - returned maximum total delegation is %f", editParams.Validator.MaximumTotalDelegation, validatorInfo.Validator.MaxTotalDelegation), verbose)
		}
	}

	if editParams.Changes.EligibilityStatus {
		if restaking.ValidatorStatus(validatorInfo.Validator.Status).String() == editParams.Validator.EligibilityStatus {
			logger.StakingLog(fmt.Sprintf("Successfully updated the eligibility status of the validator to %s", editParams.Validator.EligibilityStatus), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the eligibility status of the validator to %v - returned status is %v", editParams.Validator.EligibilityStatus, restaking.ValidatorStatus(validatorInfo.Validator.Status).String()), verbose)
		}
	}

	return editParams.Changes.TotalChanged == successfulChangeCount
}
