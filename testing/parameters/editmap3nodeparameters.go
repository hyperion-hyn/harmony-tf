package parameters

import (
	"fmt"
	"github.com/ethereum/go-ethereum/staking/types/microstaking"
	"github.com/ethereum/go-ethereum/staking/types/restaking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	"github.com/status-im/keycard-go/hexutils"
	"strings"

	sdkNetworkTypes "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/types/network"
	"github.com/hyperion-hyn/hyperion-tf/logger"
)

// EditValidatorParameters - the parameters for editing a validator
type EditMap3NodeParameters struct {
	Mode     string            `yaml:"mode"`
	Repeat   uint32            `yaml:"repeat"`
	Map3Node map3node.Map3Node `yaml:"map3Node"`

	Gas sdkNetworkTypes.Gas `yaml:"gas"`

	Changes EditValidatorChanges `yaml:"-"`

	RandomizeUniqueFields bool `yaml:"randomize_unique_fields"`
}

// EditMap3NodeChanges - keeps track of what fields have changed
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
func (editParams *EditMap3NodeParameters) Initialize() error {
	if editParams.Repeat == 0 {
		editParams.Repeat = 1
	}

	if err := editParams.Map3Node.Initialize(); err != nil {
		return err
	}

	if editParams.RandomizeUniqueFields {
		GenerateMap3NodeUniqueDetails(&editParams.Map3Node.Details)
	}

	if err := editParams.Gas.Initialize(); err != nil {
		return err
	}

	return nil
}

// DetectChanges - detects which fields have been changed during an edit validator procedure
func (editParams *EditMap3NodeParameters) DetectChanges(verbose bool) {
	if editParams.Mode != "" {
		editParams.Mode = strings.ToLower(editParams.Mode)

		switch editParams.Mode {
		case "replace_bls_key":
			editParams.Changes.ReplaceBlsKey = true
			editParams.Changes.TotalChanged++
		}
	}

	if editParams.Map3Node.Details.Name != "" {
		editParams.Changes.ValidatorName = true
		editParams.Changes.TotalChanged++
		logger.StakingLog(fmt.Sprintf("Will update the name of the validator to %s", editParams.Map3Node.Details.Name), verbose)
	}

	if editParams.Map3Node.Details.Identity != "" {
		logger.StakingLog(fmt.Sprintf("Will update the identity of the validator to %s", editParams.Map3Node.Details.Identity), verbose)
		editParams.Changes.ValidatorIdentity = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Map3Node.Details.Website != "" {
		logger.StakingLog(fmt.Sprintf("Will update the website of the validator to %s", editParams.Map3Node.Details.Website), verbose)
		editParams.Changes.ValidatorWebsite = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Map3Node.Details.SecurityContact != "" {
		logger.StakingLog(fmt.Sprintf("Will update the security contact of the validator to %s", editParams.Map3Node.Details.SecurityContact), verbose)
		editParams.Changes.ValidatorSecurityContact = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Map3Node.Details.Details != "" {
		logger.StakingLog(fmt.Sprintf("Will update the details of the validator to %s", editParams.Map3Node.Details.Details), verbose)
		editParams.Changes.ValidatorDetails = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Map3Node.RawCommission != "" {
		logger.StakingLog(fmt.Sprintf("Will update the commission rate of the validator to %f", editParams.Map3Node.Commission), verbose)
		editParams.Changes.CommissionRate = true
		editParams.Changes.TotalChanged++
	}

	if editParams.Map3Node.EligibilityStatus != "" {
		logger.StakingLog(fmt.Sprintf("Will update the eligibility status of the validator to %s", editParams.Map3Node.EligibilityStatus), verbose)
		editParams.Changes.EligibilityStatus = true
		editParams.Changes.TotalChanged++
	}
}

// EvaluateChanges - evaluates which changes have taken place and if they were successful
func (editParams *EditMap3NodeParameters) EvaluateChanges(nodeInfo microstaking.Map3NodeWrapperRPC, verbose bool) bool {
	successfulChangeCount := uint32(0)

	if editParams.Changes.ValidatorName {
		if nodeInfo.Map3Node.Description.Name == editParams.Map3Node.Details.Name {
			logger.StakingLog(fmt.Sprintf("Successfully updated the name of the validator to %s", editParams.Map3Node.Details.Name), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the name of the validator to %s - returned name is %s", editParams.Map3Node.Details.Name, nodeInfo.Map3Node.Description.Name), verbose)
		}
	}

	if editParams.Changes.ValidatorIdentity {
		if nodeInfo.Map3Node.Description.Identity == editParams.Map3Node.Details.Identity {
			logger.StakingLog(fmt.Sprintf("Successfully updated the identity of the validator to %s", editParams.Map3Node.Details.Identity), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the identity of the validator to %s - returned identity is %s", editParams.Map3Node.Details.Identity, nodeInfo.Map3Node.Description.Identity), verbose)
		}
	}

	if editParams.Changes.ValidatorWebsite {
		if nodeInfo.Map3Node.Description.Website == editParams.Map3Node.Details.Website {
			logger.StakingLog(fmt.Sprintf("Successfully updated the website of the validator to %s", editParams.Map3Node.Details.Website), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the website of the validator to %s - returned website is %s", editParams.Map3Node.Details.Website, nodeInfo.Map3Node.Description.Website), verbose)
		}
	}

	if editParams.Changes.ValidatorSecurityContact {
		if nodeInfo.Map3Node.Description.SecurityContact == editParams.Map3Node.Details.SecurityContact {
			logger.StakingLog(fmt.Sprintf("Successfully updated the security contact of the validator to %s", editParams.Map3Node.Details.SecurityContact), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the security contact of the validator to %s - returned security contact is %s", editParams.Map3Node.Details.SecurityContact, nodeInfo.Map3Node.Description.SecurityContact), verbose)
		}
	}

	if editParams.Changes.ValidatorDetails {
		if nodeInfo.Map3Node.Description.Details == editParams.Map3Node.Details.Details {
			logger.StakingLog(fmt.Sprintf("Successfully updated the details of the validator to %s", editParams.Map3Node.Details.Details), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the details of the validator to %s - returned details is %s", editParams.Map3Node.Details.Details, nodeInfo.Map3Node.Description.Details), verbose)
		}
	}

	if editParams.Changes.CommissionRate {
		if !nodeInfo.Map3Node.Commission.RateForNextPeriod.IsNil() && nodeInfo.Map3Node.Commission.RateForNextPeriod.Equal(editParams.Map3Node.Commission) {
			logger.StakingLog(fmt.Sprintf("Successfully updated the commission rate of the validator to %f", editParams.Map3Node.Commission), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the commission rate of the validator to %f - returned commission rate is %f", editParams.Map3Node.Commission, nodeInfo.Map3Node.Commission.Rate), verbose)
		}
	}

	if editParams.Changes.EligibilityStatus {
		if restaking.ValidatorStatus(nodeInfo.Map3Node.Status).String() == editParams.Map3Node.EligibilityStatus {
			logger.StakingLog(fmt.Sprintf("Successfully updated the eligibility status of the validator to %s", editParams.Map3Node.EligibilityStatus), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed to update the eligibility status of the validator to %v - returned status is %v", editParams.Map3Node.EligibilityStatus, restaking.ValidatorStatus(nodeInfo.Map3Node.Status).String()), verbose)
		}
	}

	if editParams.Changes.ReplaceBlsKey {

		editBlsKes := hexutils.BytesToHex(editParams.Map3Node.BLSKeys[0].ShardPublicKey.Key[:])
		returnBlsKeys := hexutils.BytesToHex(nodeInfo.Map3Node.NodeKeys[0][:])

		if editBlsKes == returnBlsKeys {
			logger.StakingLog(fmt.Sprintf("Successfully replace blsKey of the validator to %s", editBlsKes), verbose)
			successfulChangeCount++
		} else {
			logger.StakingLog(fmt.Sprintf("Failed replace blsKey of the validator to %v - returned status is %v", editBlsKes, returnBlsKeys), verbose)
		}

	}

	return editParams.Changes.TotalChanged == successfulChangeCount
}
