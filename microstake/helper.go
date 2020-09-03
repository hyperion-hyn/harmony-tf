package microstake

import (
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/crypto"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkCrypto "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	sdkMap3Node "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	sdkTxs "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/testing"
)

func BasicCreateMap3Node(testCase *testing.TestCase, validatorAccount *sdkAccounts.Account, senderAccount *sdkAccounts.Account, blsKeys []sdkCrypto.BLSKey) (sdkTxs.Transaction, []sdkCrypto.BLSKey, bool, error) {
	if senderAccount == nil {
		senderAccount = validatorAccount
	}

	if blsKeys == nil || len(blsKeys) == 0 {
		blsKeys = crypto.GenerateBlsKeys(testCase.StakingParameters.CreateMap3Node.BLSKeyCount, testCase.StakingParameters.CreateMap3Node.BLSSignatureMessage)
	}

	switch testCase.StakingParameters.Mode {
	case "duplicate_bls_key", "duplicateblskey":
		blsKeys = append(blsKeys, blsKeys[0])
	case "amount_larger_than_balance", "amountlargerthanbalance":
		testCase.StakingParameters.CreateMap3Node.Map3Node.Amount = testCase.StakingParameters.CreateMap3Node.Map3Node.Amount.Mul(ethCommon.NewDec(2))
	}

	if len(blsKeys) > 0 {
		for _, blsKey := range blsKeys {
			logger.StakingLog(fmt.Sprintf("Using BLS key %s to create the validator %s", blsKey.PublicKeyHex, validatorAccount.Address), testCase.Verbose)
		}
	}

	logger.TransactionLog(fmt.Sprintf("Sending create validator transaction - will wait up to %d seconds for it to finalize", testCase.StakingParameters.Timeout), testCase.Verbose)

	rawTx, err := CreateMap3Node(validatorAccount, senderAccount, &testCase.StakingParameters, blsKeys)
	if err != nil {
		return sdkTxs.Transaction{}, nil, false, err
	}

	tx := sdkTxs.ToTransaction(senderAccount.Address, senderAccount.Address, rawTx, err)
	txResultColoring := logger.ResultColoring(tx.Success, true)
	logger.TransactionLog(fmt.Sprintf("Performed create validator - address: %s - transaction hash: %s, tx successful: %s", validatorAccount.Address, tx.TransactionHash, txResultColoring), testCase.Verbose)

	rpcClient, err := config.Configuration.Network.API.RPCClient()
	validatorExists := sdkMap3Node.Exists(rpcClient, tx.ContractAddress)
	addressExistsColoring := logger.ResultColoring(validatorExists, true)
	logger.StakingLog(fmt.Sprintf("Validator with address %s exists: %s", tx.ContractAddress, addressExistsColoring), testCase.Verbose)

	return tx, blsKeys, validatorExists, nil
}
