package delegation

import (
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	hmyStaking "github.com/ethereum/go-ethereum/staking/types"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

// Delegate - delegate to a validator
func Delegate(
	keystore *keystore.KeyStore,
	account *accounts.Account,
	rpcClient *rpc.HTTPMessenger,
	chain *common.ChainID,
	delegatorAddress string,
	validatorAddress string,
	gasLimit int64,
	gasPrice ethCommon.Dec,
	nonce uint64,
	keystorePassphrase string,
	node string,
	timeout int,
) (map[string]interface{}, error) {
	payloadGenerator, err := createDelegationTransactionGenerator(delegatorAddress, validatorAddress)
	if err != nil {
		return nil, err
	}

	var logMessage string
	if network.Verbose {
		logMessage = fmt.Sprintf("Generating a new delegation transaction:\n\tDelegator Address: %s\n\tValidator Address: %s",
			delegatorAddress,
			validatorAddress,
		)
	}

	return staking.SendTx(keystore, account, rpcClient, chain, gasLimit, gasPrice, nonce, keystorePassphrase, node, timeout, payloadGenerator, logMessage)
}

func createDelegationTransactionGenerator(delegatorAddress string, validatorAddress string) (transactions.StakeMsgFulfiller, error) {
	payloadGenerator := func() (types.TransactionType, interface{}) {
		return types.Redelegate,
			hmyStaking.Redelegate{
				address.Parse(delegatorAddress),
				address.Parse(validatorAddress),
			}
	}

	return payloadGenerator, nil
}
