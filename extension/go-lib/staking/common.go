package staking

import (
	"context"
	"encoding/base64"
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	eth_hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/staking/types/restaking"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/crypto"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

// SendTx - generate the staking tx, sign it, encode the signature and send the actual tx data
func SendTx(
	keystore *keystore.KeyStore,
	account *accounts.Account,
	rpcClient *rpc.HTTPMessenger,
	chain *common.ChainID,
	gasLimit int64,
	gasPrice ethCommon.Dec,
	nonce uint64,
	keystorePassphrase string,
	node string,
	timeout int,
	payloadGenerator transactions.StakeMsgFulfiller,
	logMessage string,
) (map[string]interface{}, error) {
	if keystore == nil || account == nil {
		return nil, errors.New("keystore account can't be nil - please make sure the account you want to use exists in the keystore")
	}
	stakingTx, calculatedGasLimit, err := GenerateStakingTransaction(gasLimit, gasPrice, nonce, payloadGenerator)
	if err != nil {
		return nil, err
	}

	signedTx, err := SignStakingTransaction(keystore, account, stakingTx, chain.Value)
	if err != nil {
		return nil, err
	}

	signature, err := transactions.EncodeSignature(signedTx)
	if err != nil {
		return nil, err
	}

	if logMessage != "" {
		logMessage = fmt.Sprintf("\n[Harmony SDK]: %s - %s\n\tGas Limit: %d\n\tGas Price: %f\n\tNonce: %d\n\tSignature: %v\n",
			time.Now().Format(network.LoggingTimeFormat),
			logMessage,
			calculatedGasLimit,
			gasPrice,
			nonce,
			signature,
		)
		fmt.Println(logMessage)
	}

	receiptHash, err := SendRawStakingTransaction(rpcClient, signature)
	if err != nil {
		return nil, err
	}

	if timeout > 0 {
		hash := receiptHash.(string)
		result, _ := transactions.WaitForTxConfirmation(rpcClient, node, "staking", hash, timeout)

		if result != nil {
			return result, nil
		}
	}

	result := make(map[string]interface{})
	result["transactionHash"] = receiptHash

	return result, nil
}

// GenerateStakingTransaction - generate a staking transaction
func GenerateStakingTransaction(gasLimit int64, gasPrice ethCommon.Dec, nonce uint64, payloadGenerator transactions.StakeMsgFulfiller) (*types.Transaction, uint64, error) {
	directive, payload := payloadGenerator()
	isCreateValidator := directive == types.StakeCreateVal

	bytes, err := rlp.EncodeToBytes(payload)
	if err != nil {
		return nil, 0, err
	}

	data := base64.StdEncoding.EncodeToString(bytes)

	calculatedGasLimit, err := transactions.CalculateGasLimit(gasLimit, data, isCreateValidator)
	if err != nil {
		return nil, 0, err
	}

	// todo here add expend more staking price
	gasPrice = gasPrice.Mul(ethCommon.NewDec(params.Ether)).Quo(ethCommon.NewDec(10))

	stakingTx := types.NewTransaction(nonce, ethCommon.BigToAddress(ethCommon.Big0), big.NewInt(0), calculatedGasLimit*3, gasPrice.TruncateInt(), bytes)
	stakingTx.SetType(directive)
	return stakingTx, calculatedGasLimit, nil
}

// ProcessBlsKeys - separate bls keys to pub key and sig slices
func ProcessBlsKeys(blsKeys []crypto.BLSKey) (blsPubKeys restaking.BLSPublicKeys_, blsSigs []restaking.BLSSignature) {
	blsPubKeys = restaking.BLSPublicKeys_{
		Keys: make([]*restaking.BLSPublicKey_, len(blsKeys)),
	}
	blsSigs = make([]restaking.BLSSignature, len(blsKeys))
	return blsPubKeys, blsSigs
}

// SignStakingTransaction - sign a staking transaction
func SignStakingTransaction(keystore *keystore.KeyStore, account *accounts.Account, stakingTx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {

	signedTransaction, err := keystore.SignTx(*account, stakingTx, chainID)
	if err != nil {
		return nil, err
	}

	return signedTransaction, nil
}

// SendRawStakingTransaction - send the raw staking tx to the RPC endpoint
func SendRawStakingTransaction(rpcClient *rpc.HTTPMessenger, signature *string) (interface{}, error) {
	rawTxBytes, err := eth_hexutil.Decode(*signature)
	tx := new(types.Transaction)
	rlp.DecodeBytes(rawTxBytes, &tx)
	err = rpcClient.GetClient().SendTransaction(context.Background(), tx)
	if err != nil {
		return nil, err
	}

	time.Sleep(10 * time.Second)

	fmt.Printf("tx sent: %s", tx.Hash().Hex())

	return tx.Hash().Hex(), nil
}

// NumericDecToBigIntAmount - convert a ethCommon.Dec amount to a converted big.Int amount
func NumericDecToBigIntAmount(amount ethCommon.Dec) (bigAmount *big.Int) {
	if !amount.IsNil() {
		amount = amount.Mul(transactions.OneAsDec)
		bigAmount = amount.RoundInt()
	}

	return bigAmount
}
