package parameters

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/staking/types/microstaking"
	"github.com/ethereum/go-ethereum/staking/types/restaking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	"github.com/status-im/keycard-go/hexutils"
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
	Map3Node  map3node.Map3Node      `yaml:"map3Node"`

	Gas sdkNetworkTypes.Gas `yaml:"gas"`

	Changes         EditValidatorChanges `yaml:"-"`
	Map3NodeChanges EditMap3NodeChanges  `yaml:"-"`

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
	ReplaceBlsKey            bool   `yaml:"-"`
	TotalChanged             uint32 `yaml:"-"`
}
type EditMap3NodeChanges struct {
	Map3NodeName            bool   `yaml:"-"`
	Map3NodeIdentity        bool   `yaml:"-"`
	Map3NodeWebsite         bool   `yaml:"-"`
	Map3NodeSecurityContact bool   `yaml:"-"`
	Map3NodeDetails         bool   `yaml:"-"`
	CommissionRate          bool   `yaml:"-"`
	ReplaceBlsKey           bool   `yaml:"-"`
	TotalChanged            uint32 `yaml:"-"`
}

// Initialize - initializes the edit staking parameters
func (editParams *EditValidatorParameters) Initialize() error {
	if editParams.Repeat == 0 {
		editParams.Repeat = 1
	}

	if err := editParams.Validator.Initialize(); err != nil {
		return err
	}
	if err := editParams.Map3Node.Initialize(); err != nil {
		return err
	}

	if editParams.RandomizeUniqueFields {
		GenerateUniqueDetails(&editParams.Validator.Details)
		GenerateMap3NodeUniqueDetails(&editParams.Map3Node.Details)
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

		switch editParams.Mode {
		case "replace_bls_key":
			editParams.Changes.ReplaceBlsKey = true
			editParams.Changes.TotalChanged++
		}
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
func (editParams *EditValidatorParameters) EvaluateChanges(validatorInfo restaking.PlainValidatorWrapper, verbose bool) bool {
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
		var actualMaxTotalDelegation common.Dec

		if validatorInfo.Validator.MaxTotalDelegation != nil {
			actualMaxTotalDelegation = common.NewDecFromBigInt(validatorInfo.Validator.MaxTotalDelegation).QuoInt64(params.Ether)
		} else {
			actualMaxTotalDelegation = common.NewDec(0)
		}

		if !actualMaxTotalDelegation.Equal(common.NewDec(0)) && actualMaxTotalDelegation.Equal(editParams.Validator.MaximumTotalDelegation) {
			logger.StakingLog(fmt.Sprintf("Successfully updated the maximum total delegation of the validator to %f", editParams.Validator.MaximumTotalDelegation), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the maximum total delegation of the validator to %f - returned maximum total delegation is %f", editParams.Validator.MaximumTotalDelegation, actualMaxTotalDelegation), verbose)
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

	if editParams.Changes.ReplaceBlsKey {

		editBlsKes := hexutils.BytesToHex(editParams.Validator.BLSKeys[0].ShardPublicKey.Key[:])
		returnBlsKeys := hexutils.BytesToHex(validatorInfo.Validator.SlotPubKeys[0].Key[:])

		if editBlsKes == returnBlsKeys {
			logger.StakingLog(fmt.Sprintf("Successfully replace blsKey of the validator to %s", editBlsKes), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed replace blsKey of the validator to %v - returned status is %v", editBlsKes, returnBlsKeys), verbose)
		}

	}

	return editParams.Changes.TotalChanged == successfulChangeCount
}

// DetectChanges - detects which fields have been changed during an edit validator procedure
func (editParams *EditValidatorParameters) DetectMap3NodeChanges(verbose bool) {
	if editParams.Mode != "" {
		editParams.Mode = strings.ToLower(editParams.Mode)

		switch editParams.Mode {
		case "replace_bls_key":
			editParams.Map3NodeChanges.ReplaceBlsKey = true
			editParams.Map3NodeChanges.TotalChanged++
		}
	}

	if editParams.Map3Node.Details.Name != "" {
		editParams.Map3NodeChanges.Map3NodeName = true
		editParams.Map3NodeChanges.TotalChanged++
		logger.StakingLog(fmt.Sprintf("Will update the name of the validator to %s", editParams.Map3Node.Details.Name), verbose)
	}

	if editParams.Map3Node.Details.Identity != "" {
		logger.StakingLog(fmt.Sprintf("Will update the identity of the validator to %s", editParams.Map3Node.Details.Identity), verbose)
		editParams.Map3NodeChanges.Map3NodeIdentity = true
		editParams.Map3NodeChanges.TotalChanged++
	}

	if editParams.Map3Node.Details.Website != "" {
		logger.StakingLog(fmt.Sprintf("Will update the website of the validator to %s", editParams.Map3Node.Details.Website), verbose)
		editParams.Map3NodeChanges.Map3NodeWebsite = true
		editParams.Map3NodeChanges.TotalChanged++
	}

	if editParams.Map3Node.Details.SecurityContact != "" {
		logger.StakingLog(fmt.Sprintf("Will update the security contact of the validator to %s", editParams.Map3Node.Details.SecurityContact), verbose)
		editParams.Map3NodeChanges.Map3NodeSecurityContact = true
		editParams.Map3NodeChanges.TotalChanged++
	}

	if editParams.Map3Node.Details.Details != "" {
		logger.StakingLog(fmt.Sprintf("Will update the details of the validator to %s", editParams.Map3Node.Details.Details), verbose)
		editParams.Map3NodeChanges.Map3NodeDetails = true
		editParams.Map3NodeChanges.TotalChanged++
	}

	if editParams.Map3Node.RawCommission != "" {
		logger.StakingLog(fmt.Sprintf("Will update the commission rate of the validator to %f", editParams.Map3Node.Commission), verbose)
		editParams.Map3NodeChanges.CommissionRate = true
		editParams.Map3NodeChanges.TotalChanged++
	}

	//if editParams.Map3Node.EligibilityStatus != "" {
	//	logger.StakingLog(fmt.Sprintf("Will update the eligibility status of the validator to %s", editParams.Map3Node.EligibilityStatus), verbose)
	//	editParams.Map3NodeChanges.EligibilityStatus = true
	//	editParams.Map3NodeChanges.TotalChanged++
	//}
}

// EvaluateChanges - evaluates which changes have taken place and if they were successful
func (editParams *EditValidatorParameters) EvaluateMap3NodeChanges(nodeInfo microstaking.PlainMap3NodeWrapper, verbose bool) bool {
	successfulChangeCount := uint32(0)

	if editParams.Map3NodeChanges.Map3NodeName {
		if nodeInfo.Map3Node.Description.Name == editParams.Map3Node.Details.Name {
			logger.StakingLog(fmt.Sprintf("Successfully updated the name of the validator to %s", editParams.Map3Node.Details.Name), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the name of the validator to %s - returned name is %s", editParams.Map3Node.Details.Name, nodeInfo.Map3Node.Description.Name), verbose)
		}
	}

	if editParams.Map3NodeChanges.Map3NodeIdentity {
		if nodeInfo.Map3Node.Description.Identity == editParams.Map3Node.Details.Identity {
			logger.StakingLog(fmt.Sprintf("Successfully updated the identity of the validator to %s", editParams.Map3Node.Details.Identity), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the identity of the validator to %s - returned identity is %s", editParams.Map3Node.Details.Identity, nodeInfo.Map3Node.Description.Identity), verbose)
		}
	}

	if editParams.Map3NodeChanges.Map3NodeWebsite {
		if nodeInfo.Map3Node.Description.Website == editParams.Map3Node.Details.Website {
			logger.StakingLog(fmt.Sprintf("Successfully updated the website of the validator to %s", editParams.Map3Node.Details.Website), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the website of the validator to %s - returned website is %s", editParams.Map3Node.Details.Website, nodeInfo.Map3Node.Description.Website), verbose)
		}
	}

	if editParams.Map3NodeChanges.Map3NodeSecurityContact {
		if nodeInfo.Map3Node.Description.SecurityContact == editParams.Map3Node.Details.SecurityContact {
			logger.StakingLog(fmt.Sprintf("Successfully updated the security contact of the validator to %s", editParams.Map3Node.Details.SecurityContact), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the security contact of the validator to %s - returned security contact is %s", editParams.Map3Node.Details.SecurityContact, nodeInfo.Map3Node.Description.SecurityContact), verbose)
		}
	}

	if editParams.Map3NodeChanges.Map3NodeDetails {
		if nodeInfo.Map3Node.Description.Details == editParams.Map3Node.Details.Details {
			logger.StakingLog(fmt.Sprintf("Successfully updated the details of the validator to %s", editParams.Map3Node.Details.Details), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the details of the validator to %s - returned details is %s", editParams.Map3Node.Details.Details, nodeInfo.Map3Node.Description.Details), verbose)
		}
	}

	if editParams.Map3NodeChanges.CommissionRate {
		if !nodeInfo.Map3Node.Commission.RateForNextPeriod.IsNil() && nodeInfo.Map3Node.Commission.RateForNextPeriod.Equal(editParams.Map3Node.Commission) {
			logger.StakingLog(fmt.Sprintf("Successfully updated the commission rate of the validator to %f", editParams.Map3Node.Commission), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the commission rate of the validator to %f - returned commission rate is %f", editParams.Map3Node.Commission, nodeInfo.Map3Node.Commission.Rate), verbose)
		}
	}

	//if editParams.Map3NodeChanges.EligibilityStatus {
	//	if restaking.ValidatorStatus(nodeInfo.Map3Node.Status).String() == editParams.Map3Node.EligibilityStatus {
	//		logger.StakingLog(fmt.Sprintf("Successfully updated the eligibility status of the validator to %s", editParams.Map3Node.EligibilityStatus), verbose)
	//		successfulChangeCount++
	//	} else {
	//		logger.StakingLog(fmt.Sprintf("Failed to update the eligibility status of the validator to %v - returned status is %v", editParams.Map3Node.EligibilityStatus, restaking.ValidatorStatus(nodeInfo.Map3Node.Status).String()), verbose)
	//	}
	//}

	if editParams.Map3NodeChanges.ReplaceBlsKey {

		editBlsKes := hexutils.BytesToHex(editParams.Map3Node.BLSKeys[0].ShardPublicKey.Key[:])
		returnBlsKeys := hexutils.BytesToHex(nodeInfo.Map3Node.NodeKeys[0].Key[:])

		if editBlsKes == returnBlsKeys {
			logger.StakingLog(fmt.Sprintf("Successfully replace blsKey of the validator to %s", editBlsKes), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed replace blsKey of the validator to %v - returned status is %v", editBlsKes, returnBlsKeys), verbose)
		}

	}

	return editParams.Map3NodeChanges.TotalChanged == successfulChangeCount
}
