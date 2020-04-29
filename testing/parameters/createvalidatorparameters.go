package parameters

import (
	sdkCrypto "github.com/harmony-one/go-lib/crypto"
	sdkValidator "github.com/harmony-one/go-lib/staking/validator"
)

// CreateValidatorParameters - represents the validator details
type CreateValidatorParameters struct {
	Validator             sdkValidator.Validator `yaml:"validator"`
	BLSKeyCount           int                    `yaml:"bls_key_count"`
	BLSSignatureMessage   string                 `yaml:"bls_signature_message"`
	RandomizeUniqueFields bool                   `yaml:"randomize_unique_fields"`
}

// Initialize - initializes and converts values
func (params *CreateValidatorParameters) Initialize() error {
	if err := params.Validator.Initialize(); err != nil {
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

	if params.RandomizeUniqueFields {
		GenerateUniqueDetails(&params.Validator.Details)
	}

	return nil
}
