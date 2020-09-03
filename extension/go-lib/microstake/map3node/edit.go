package map3node

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	stakingCommon "github.com/ethereum/go-ethereum/staking/types/common"
	"github.com/ethereum/go-ethereum/staking/types/microstaking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

// Edit - edits the details for an existing validator
func Edit(
	keystore *keystore.KeyStore,
	account *accounts.Account,
	rpcClient *rpc.HTTPMessenger,
	chain *common.ChainID,
	map3NodeAddress string,
	description microstaking.Description_,
	blsKeyToRemove *crypto.BLSKey,
	blsKeyToAdd *crypto.BLSKey,
	gasLimit int64,
	gasPrice ethCommon.Dec,
	nonce uint64,
	keystorePassphrase string,
	node string,
	timeout int,
) (map[string]interface{}, error) {
	payloadGenerator, err := editTransactionGenerator(address.Parse(map3NodeAddress), account.Address, description, blsKeyToRemove, blsKeyToAdd)
	if err != nil {
		return nil, err
	}

	var logMessage string
	if network.Verbose {
		logMessage = fmt.Sprintf("Generating a new edit validator transaction:\n\tValidator Address: %s\n\tValidator Name: %s\n\tValidator Identity: %s\n\tValidator Website: %s\n\tValidator Security Contact: %s\n\tValidator Details: %s\n\tRemove BLS key: %v\n\tAdd BLS key: %v\n\t",
			map3NodeAddress,
			description.Name,
			description.Identity,
			description.Website,
			description.SecurityContact,
			description.Details,
			blsKeyToRemove,
			blsKeyToAdd,
		)
	}

	return staking.SendTx(keystore, account, rpcClient, chain, gasLimit, gasPrice, nonce, keystorePassphrase, node, timeout, payloadGenerator, logMessage)
}

func editTransactionGenerator(
	map3NodeAddress ethCommon.Address,
	operatorAddress ethCommon.Address,
	microstakingDescription microstaking.Description_,
	blsKeyToRemove *crypto.BLSKey,
	blsKeyToAdd *crypto.BLSKey,
) (transactions.StakeMsgFulfiller, error) {
	var shardBlsKeyToRemove *microstaking.BLSPublicKey_
	if blsKeyToRemove != nil {
		shardBlsKeyToRemove = blsKeyToRemove.NodePublicKey
	}

	var shardBlsKeyToAdd *microstaking.BLSPublicKey_
	var shardBlsKeyToAddSig *stakingCommon.BLSSignature
	if blsKeyToAdd != nil {
		shardBlsKeyToAdd = blsKeyToAdd.NodePublicKey
		shardBlsKeyToAddSig = blsKeyToAdd.ShardSignature
	}

	payloadGenerator := func() (types.TransactionType, interface{}) {
		return types.EditMap3, microstaking.EditMap3Node{
			Map3NodeAddress: map3NodeAddress,
			OperatorAddress: operatorAddress,
			Description:     microstakingDescription,
			NodeKeyToRemove: shardBlsKeyToRemove,
			NodeKeyToAdd:    shardBlsKeyToAdd,
			NodeKeyToAddSig: shardBlsKeyToAddSig,
		}
	}

	return payloadGenerator, nil
}
