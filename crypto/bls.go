package crypto

import (
	"math/big"

	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkCrypto "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
)

// GenerateBlsKeys - generates a set of bls keys given a count
func GenerateBlsKeys(count int, shardID uint32, message string) (blsKeys []sdkCrypto.BLSKey) {
	if count > 0 {
		for i := 0; i < count; i++ {
			var blsKey sdkCrypto.BLSKey
			var err error

			blsKey, err = GenerateBlsKey(shardID, message)

			//fmt.Println(fmt.Sprintf("Generated bls key - private key: %s, public key: %s, shard public key: %v, shard signature: %v", blsKey.PrivateKeyHex, blsKey.PublicKeyHex, blsKey.ShardPublicKey, blsKey.ShardSignature))

			if err == nil {
				blsKeys = append(blsKeys, blsKey)
			}
		}
	}

	return blsKeys
}

// GenerateBlsKey - generates a new bls key
func GenerateBlsKey(shardID uint32, message string) (blsKey sdkCrypto.BLSKey, err error) {
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
	bigShardCount := big.NewInt(int64(config.Configuration.Network.Shards))
	resolvedShardID := int(new(big.Int).Mod(blsKey.ShardPublicKey.Big(), bigShardCount).Int64())
	return (int(desiredShardID) == resolvedShardID)
}
