package crypto

import (
	sdkCrypto "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	"math/big"
)

// GenerateBlsKeys - generates a set of bls keys given a count
func GenerateBlsKeys(count int, message string) (blsKeys []sdkCrypto.BLSKey) {
	if count > 0 {
		for i := 0; i < count; i++ {
			var blsKey sdkCrypto.BLSKey
			var err error

			blsKey, err = GenerateBlsKey(message)

			//fmt.Println(fmt.Sprintf("Generated bls key - private key: %s, public key: %s, shard public key: %v, shard signature: %v", blsKey.PrivateKeyHex, blsKey.PublicKeyHex, blsKey.ShardPublicKey, blsKey.ShardSignature))

			if err == nil {
				blsKeys = append(blsKeys, blsKey)
			}
		}
	}

	return blsKeys
}

// GenerateBlsKey - generates a new bls key
func GenerateBlsKey(message string) (blsKey sdkCrypto.BLSKey, err error) {

	var shardID uint32 = 0

	for {
		blsKey, err = sdkCrypto.GenerateBlsKey(message)
		if err != nil {
			return sdkCrypto.BLSKey{}, err
		}

		if blsKeyMatchesShardID(blsKey, shardID) {
			break
		}
	}

	return blsKey, nil
}

func blsKeyMatchesShardID(blsKey sdkCrypto.BLSKey, desiredShardID uint32) bool {
	bigShardCount := big.NewInt(int64(2))
	resolvedShardID := int(new(big.Int).Mod(blsKey.ShardPublicKey.Big(), bigShardCount).Int64())
	return (int(desiredShardID) == resolvedShardID)
}
