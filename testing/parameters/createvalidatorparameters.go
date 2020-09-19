package parameters

import (
	sdkCrypto "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	sdkValidator "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking/validator"
)

// CreateValidatorParameters - represents the validator details
type CreateValidatorParameters struct {
	Validator             sdkValidator.Validator `yaml:"validator"`
	Map3Node              map3node.Map3Node      `yaml:"map3Node"`
	BLSKeyCount           int                    `yaml:"bls_key_count"`
	BLSSignatureMessage   string                 `yaml:"bls_signature_message"`
	RandomizeUniqueFields bool                   `yaml:"randomize_unique_fields"`
}

// Initialize - initializes and converts values
func (params *CreateValidatorParameters) Initialize() error {
	if err := params.Validator.Initialize(); err != nil {
		return err
	}
	if err := params.Map3Node.Initialize(); err != nil {
		return err
	}

	if params.BLSKeyCount < 0 {
		params.BLSKeyCount = 1
	}

	if len(params.Validator.BLSKeys) == 0 && params.BLSKeyCount > 0 {
		for i := 0; i < params.BLSKeyCount; i++ {
			blsKey, err := sdkCrypto.GenerateBlsKey(params.BLSSignatureMessage)
			if err != nil {
				return err
			}

			params.Validator.BLSKeys = append(params.Validator.BLSKeys, blsKey)
		}
	}
	if len(params.Map3Node.BLSKeys) == 0 && params.BLSKeyCount > 0 {
		for i := 0; i < params.BLSKeyCount; i++ {
			blsKey, err := sdkCrypto.GenerateBlsKey(params.BLSSignatureMessage)
			if err != nil {
				return err
			}

			params.Map3Node.BLSKeys = append(params.Map3Node.BLSKeys, blsKey)
		}
	}

	if params.RandomizeUniqueFields {
		GenerateUniqueDetails(&params.Validator.Details)
		GenerateMap3NodeUniqueDetails(&params.Map3Node.Details)
	}

	return nil
}
