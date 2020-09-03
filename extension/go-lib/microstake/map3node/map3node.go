package map3node

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/staking/types/microstaking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/pkg/errors"
)

// Map3Node - represents the map3Node details
type Map3Node struct {
	Map3Address   string `yaml:"-"`
	Account       *accounts.Account
	Details       Map3NodeDetails `yaml:"details"`
	RawCommission string          `yaml:"commission"`
	Commission    ethCommon.Dec   `yaml:"-"`
	BLSKeys       []crypto.BLSKey `yaml:"-"`
	Exists        bool

	RawAmount string        `yaml:"amount"`
	Amount    ethCommon.Dec `yaml:"-"`

	EligibilityStatus string `yaml:"eligibility-status"`
}

// Map3NodeDetails - represents the map3Node details
type Map3NodeDetails struct {
	Name            string `yaml:"name"`
	Identity        string `yaml:"identity"`
	Website         string `yaml:"website"`
	SecurityContact string `yaml:"security_contact"`
	Details         string `yaml:"details"`
}

// Initialize - initializes and converts values for a given map3Node
func (map3Node *Map3Node) Initialize() error {

	if map3Node.RawAmount != "" {
		decAmount, err := common.NewDecFromString(map3Node.RawAmount)
		if err != nil {
			return errors.Wrapf(err, "Map3Node: Amount")
		}
		map3Node.Amount = decAmount
	}
	if map3Node.RawCommission != "" {
		decCommission, err := common.NewDecFromString(map3Node.RawCommission)
		if err != nil {
			return errors.Wrapf(err, "Map3Node: Commission")
		}
		map3Node.Commission = decCommission
	}

	return nil
}

func (map3Node *Map3Node) ToMicroStakeDescription() microstaking.Description_ {
	return microstaking.Description_{
		Name:            map3Node.Details.Name,
		Identity:        map3Node.Details.Identity,
		Website:         map3Node.Details.Website,
		SecurityContact: map3Node.Details.SecurityContact,
		Details:         map3Node.Details.Details,
	}
}
