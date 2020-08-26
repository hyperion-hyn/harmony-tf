package validator

import (
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	restaking "github.com/ethereum/go-ethereum/staking/types/restaking"
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
	description restaking.Description_,
	commissionRates restaking.CommissionRates_,
	maximumTotalDelegation ethCommon.Dec,
	blsKeys []crypto.BLSKey,
	gasLimit int64,
	gasPrice ethCommon.Dec,
	nonce uint64,
	keystorePassphrase string,
	node string,
	timeout int,
) (map[string]interface{}, error) {
	payloadGenerator, err := createTransactionGenerator(operatorAddress, description, commissionRates, maximumTotalDelegation, blsKeys)
	if err != nil {
		return nil, err
	}

	var logMessage string
	if network.Verbose {
		logMessage = fmt.Sprintf("Generating a new create validator transaction:\n\tValidator Address: %s\n\tValidator Name: %s\n\tValidator Identity: %s\n\tValidator Website: %s\n\tValidator Security Contact: %s\n\tValidator Details: %s\n\tCommission Rate: %f\n\tCommission Max Rate: %f\n\tCommission Max Change Rate: %d\n\tMaximum Total Delegation: %f\n\tBls Public Keys: %v",
			operatorAddress,
			description.Name,
			description.Identity,
			description.Website,
			description.SecurityContact,
			description.Details,
			commissionRates.Rate,
			commissionRates.MaxRate,
			commissionRates.MaxChangeRate,
			maximumTotalDelegation,
			blsKeys,
		)
	}

	return staking.SendTx(keystore, account, rpcClient, chain, gasLimit, gasPrice, nonce, keystorePassphrase, node, timeout, payloadGenerator, logMessage)
}

func createTransactionGenerator(
	validatorAddress string,
	stakingDescription restaking.Description_,
	stakingCommissionRates restaking.CommissionRates_,
	maximumTotalDelegation ethCommon.Dec,
	blsKeys []crypto.BLSKey,
) (transactions.StakeMsgFulfiller, error) {
	//blsPubKeys, blsSigs := staking.ProcessBlsKeys(blsKeys)
	bigMaximumTotalDelegation := staking.NumericDecToBigIntAmount(maximumTotalDelegation)

	var slotPubKey restaking.BLSPublicKey_

	var slotKeySig restaking.BLSSignature

	if len(blsKeys) == 0 {
		slotPubKey = restaking.BLSPublicKey_{Key: [48]byte{}}
		slotKeySig = [restaking.BLSSignatureSizeInBytes]byte{}
	} else {
		blsKey := blsKeys[0]
		slotPubKey = *blsKey.ShardPublicKey
		slotKeySig = *blsKey.ShardSignature
	}

	//println(blsSigs) // todo need remove
	payloadGenerator := func() (types.TransactionType, interface{}) {
		return types.StakeCreateVal, restaking.CreateValidator{
			OperatorAddress:    address.Parse(validatorAddress),
			Description:        stakingDescription,
			CommissionRates:    stakingCommissionRates,
			MaxTotalDelegation: bigMaximumTotalDelegation,
			SlotPubKey:         slotPubKey,
			SlotKeySig:         slotKeySig,
			//SlotKeySig: nil, // todo need revert
		}
	}

	return payloadGenerator, nil
}
