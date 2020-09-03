package parameters

import (
	sdkCrypto "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
)

// CreateValidatorParameters - represents the validator details
type CreateMap3NodeParameters struct {
	Map3Node              map3node.Map3Node `yaml:"map3Node"`
	BLSKeyCount           int               `yaml:"bls_key_count"`
	BLSSignatureMessage   string            `yaml:"bls_signature_message"`
	RandomizeUniqueFields bool              `yaml:"randomize_unique_fields"`
}

// Initialize - initializes and converts values
func (params *CreateMap3NodeParameters) Initialize() error {
	if err := params.Map3Node.Initialize(); err != nil {
		return err
	}

	if params.BLSKeyCount < 0 {
		params.BLSKeyCount = 1
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
		GenerateMap3NodeUniqueDetails(&params.Map3Node.Details)
	}

	return nil
}
