package validator

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	restaking "github.com/ethereum/go-ethereum/staking/types/restaking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
	"strings"
)

// Edit - edits the details for an existing validator
func Edit(
	keystore *keystore.KeyStore,
	account *accounts.Account,
	rpcClient *rpc.HTTPMessenger,
	chain *common.ChainID,
	validatorAddress string,
	description restaking.Description_,
	commissionRate *ethCommon.Dec,
	minimumSelfDelegation ethCommon.Dec,
	maximumTotalDelegation ethCommon.Dec,
	blsKeyToRemove *crypto.BLSKey,
	blsKeyToAdd *crypto.BLSKey,
	status string,
	gasLimit int64,
	gasPrice ethCommon.Dec,
	nonce uint64,
	keystorePassphrase string,
	node string,
	timeout int,
) (map[string]interface{}, error) {
	statusEnum := determineEposStatus(status)

	payloadGenerator, err := editTransactionGenerator(validatorAddress, description, *commissionRate, maximumTotalDelegation, blsKeyToRemove, blsKeyToAdd, statusEnum)
	if err != nil {
		return nil, err
	}

	var logMessage string
	if network.Verbose {
		logMessage = fmt.Sprintf("Generating a new edit validator transaction:\n\tValidator Address: %s\n\tValidator Name: %s\n\tValidator Identity: %s\n\tValidator Website: %s\n\tValidator Security Contact: %s\n\tValidator Details: %s\n\tCommission Rate: %v\n\tMinimum Self Delegation: %f\n\tMaximum Total Delegation: %f\n\tRemove BLS key: %v\n\tAdd BLS key: %v\n\tStatus: %v",
			validatorAddress,
			description.Name,
			description.Identity,
			description.Website,
			description.SecurityContact,
			description.Details,
			commissionRate,
			minimumSelfDelegation,
			maximumTotalDelegation,
			blsKeyToRemove,
			blsKeyToAdd,
			statusEnum,
		)
	}

	return staking.SendTx(keystore, account, rpcClient, chain, gasLimit, gasPrice, nonce, keystorePassphrase, node, timeout, payloadGenerator, logMessage)
}

func determineEposStatus(status string) (statusEnum restaking.ValidatorStatus) {
	switch strings.ToLower(status) {
	case "active":
		return restaking.Active
	case "inactive":
		return restaking.Inactive
	default:
		return restaking.Nil
	}
}

func editTransactionGenerator(
	validatorAddress string,
	stakingDescription restaking.Description_,
	commissionRate ethCommon.Dec,
	maximumTotalDelegation ethCommon.Dec,
	blsKeyToRemove *crypto.BLSKey,
	blsKeyToAdd *crypto.BLSKey,
	statusEnum restaking.ValidatorStatus,
) (transactions.StakeMsgFulfiller, error) {
	var shardBlsKeyToRemove *restaking.BLSPublicKey_
	if blsKeyToRemove != nil {
		shardBlsKeyToRemove = blsKeyToRemove.ShardPublicKey
	}

	var shardBlsKeyToAdd *restaking.BLSPublicKey_
	var shardBlsKeyToAddSig *restaking.BLSSignature
	if blsKeyToAdd != nil {
		shardBlsKeyToAdd = blsKeyToAdd.ShardPublicKey
		shardBlsKeyToAddSig = blsKeyToAdd.ShardSignature
	}

	bigMaximumTotalDelegation := staking.NumericDecToBigIntAmount(maximumTotalDelegation)

	println(shardBlsKeyToAddSig) // todo need remove
	payloadGenerator := func() (types.TransactionType, interface{}) {
		return types.StakeEditVal, restaking.EditValidator{
			ValidatorAddress:   address.Parse(validatorAddress),
			Description:        stakingDescription,
			CommissionRate:     &commissionRate,
			MaxTotalDelegation: bigMaximumTotalDelegation,
			SlotKeyToRemove:    shardBlsKeyToRemove,
			SlotKeyToAdd:       shardBlsKeyToAdd,
			//SlotKeyToAddSig:    shardBlsKeyToAddSig,
			SlotKeyToAddSig: nil, // todo need remove
			EPOSStatus:      statusEnum,
		}
	}

	return payloadGenerator, nil
}
