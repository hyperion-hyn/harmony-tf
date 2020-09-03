package map3node

import (
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	stakingCommon "github.com/ethereum/go-ethereum/staking/types/common"
	"github.com/ethereum/go-ethereum/staking/types/microstaking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

// Create - creates a validator
func Create(
	keystore *keystore.KeyStore,
	account *accounts.Account,
	rpcClient *rpc.HTTPMessenger,
	chain *common.ChainID,
	operatorAddress string,
	description microstaking.Description_,
	commissionRates ethCommon.Dec,
	blsKeys []crypto.BLSKey,
	amount ethCommon.Dec,
	gasLimit int64,
	gasPrice ethCommon.Dec,
	nonce uint64,
	keystorePassphrase string,
	node string,
	timeout int,
) (map[string]interface{}, error) {
	payloadGenerator, err := createTransactionGenerator(operatorAddress, description, commissionRates, blsKeys, amount)
	if err != nil {
		return nil, err
	}

	var logMessage string
	if network.Verbose {
		logMessage = fmt.Sprintf("Generating a new create map3Node transaction:\n\tMap3NodeAccount Address: %s\n\tValidator Name: %s\n\tValidator Identity: %s\n\tValidator Website: %s\n\tValidator Security Contact: %s\n\tValidator Details: %s\n\tCommission Rate: %f\n\tAmount: %f\n\tBls Public Keys: %v",
			operatorAddress,
			description.Name,
			description.Identity,
			description.Website,
			description.SecurityContact,
			description.Details,
			commissionRates,
			amount,
			blsKeys,
		)
	}

	return staking.SendTx(keystore, account, rpcClient, chain, gasLimit, gasPrice, nonce, keystorePassphrase, node, timeout, payloadGenerator, logMessage)
}

func createTransactionGenerator(
	validatorAddress string,
	stakingDescription microstaking.Description_,
	stakingCommissionRates ethCommon.Dec,
	blsKeys []crypto.BLSKey,
	amount ethCommon.Dec,
) (transactions.StakeMsgFulfiller, error) {
	//blsPubKeys, blsSigs := staking.ProcessBlsKeys(blsKeys)

	var slotPubKey microstaking.BLSPublicKey_

	var slotKeySig stakingCommon.BLSSignature

	if len(blsKeys) == 0 {
		slotPubKey = microstaking.BLSPublicKey_{Key: [48]byte{}}
		slotKeySig = [stakingCommon.BLSSignatureSizeInBytes]byte{}
	} else {
		blsKey := blsKeys[0]
		slotPubKey = *blsKey.NodePublicKey
		slotKeySig = *blsKey.ShardSignature
	}

	bigAmount := staking.NumericDecToBigIntAmount(amount)

	//println(blsSigs) // todo need remove
	payloadGenerator := func() (types.TransactionType, interface{}) {
		return types.CreateMap3, microstaking.CreateMap3Node{
			OperatorAddress: address.Parse(validatorAddress),
			Description:     stakingDescription,
			Commission:      stakingCommissionRates,
			NodePubKey:      slotPubKey,
			NodeKeySig:      slotKeySig,
			Amount:          bigAmount,
			//SlotKeySig: nil, // todo need revert
		}
	}

	return payloadGenerator, nil
}
