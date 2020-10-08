package transactions

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	eth_hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	eth_rlp "github.com/ethereum/go-ethereum/rlp"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/rpc"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	goSdkRPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/transaction"
)

// Copied from harmony-one/harmony/internal/params/protocol_params.go
const (
	// TxGas ...
	TxGas uint64 = 21000 // Per transaction not creating a contract. NOTE: Not payable on data of calls between transactions.
	// TxGasContractCreation ...
	TxGasContractCreation uint64 = 53000 // Per transaction that creates a contract. NOTE: Not payable on data of calls between transactions.
	// TxGasValidatorCreation ...
	TxGasValidatorCreation uint64 = 5300000 // Per transaction that creates a new validator. NOTE: Not payable on data of calls between transactions.
)

var (
	// NanoAsDec - Nano denomination in ethCommon.Dec
	NanoAsDec = ethCommon.NewDec(params.GWei)
	// OneAsDec - One denomination in ethCommon.Dec
	OneAsDec = ethCommon.NewDec(params.Ether)
)

// SendTransaction - send transactions
func SendTransaction(keystore *keystore.KeyStore, account *accounts.Account, rpcClient *goSdkRPC.HTTPMessenger, chain *common.ChainID, fromAddress string, toAddress string, amount ethCommon.Dec, gasLimit int64, gasPrice ethCommon.Dec, nonce uint64, inputData string, keystorePassphrase string, node string, timeout int) (map[string]interface{}, error) {
	if keystore == nil || account == nil {
		return nil, errors.New("keystore account can't be nil - please make sure the account you want to use exists in the keystore")
	}

	signedTx, err := GenerateAndSignTransaction(
		keystore,
		account,
		chain,
		fromAddress,
		toAddress,
		amount,
		gasLimit,
		gasPrice,
		nonce,
		inputData,
	)
	if err != nil {
		return nil, err
	}

	signature, err := EncodeSignature(signedTx)
	if err != nil {
		return nil, err
	}

	if network.Verbose {
		fmt.Printf("\n[Harmony SDK]: %s - signed transaction using chain: %s (id: %d), signature: %v\n", time.Now().Format(network.LoggingTimeFormat), chain.Name, chain.Value, signature)
		json, _ := signedTx.MarshalJSON()
		fmt.Printf("%s\n", common.JSONPrettyFormat(string(json)))
		fmt.Printf("\n[Harmony SDK]: %s - sending transaction using node: %s, chain: %s (id: %d), signature: %v, timeout: %d\n\n", time.Now().Format(network.LoggingTimeFormat), node, chain.Name, chain.Value, signature, timeout)
	}

	receiptHash, err := SendRawTransaction(rpcClient, signature)
	if err != nil {
		return nil, err
	}

	if timeout > 0 {
		hash := receiptHash.(string)
		result, err := WaitForTxConfirmation(rpcClient, node, "transaction", hash, timeout)
		if err != nil {
			return nil, err
		}

		if result != nil {
			return result, nil
		}
	}

	result := make(map[string]interface{})
	result["transactionHash"] = receiptHash

	return result, nil
}

// GenerateAndSignTransaction - generates and signs a transaction based on the supplied tx params and keystore/account
func GenerateAndSignTransaction(
	keystore *keystore.KeyStore,
	account *accounts.Account,
	chain *common.ChainID,
	fromAddress string,
	toAddress string,
	amount ethCommon.Dec,
	gasLimit int64,
	gasPrice ethCommon.Dec,
	nonce uint64,
	inputData string,
) (tx *types.Transaction, err error) {
	generatedTx, err := GenerateTransaction(fromAddress, toAddress, amount, gasLimit, gasPrice, nonce, inputData)
	if err != nil {
		return nil, err
	}

	tx, err = SignTransaction(keystore, account, generatedTx, chain.Value)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// GenerateTransaction - generate a new transaction
func GenerateTransaction(
	fromAddress string,
	toAddress string,
	amount ethCommon.Dec,
	gasLimit int64,
	gasPrice ethCommon.Dec,
	nonce uint64,
	inputData string,
) (tx *transaction.Transaction, err error) {
	calculatedGasLimit, err := CalculateGasLimit(gasLimit, inputData, false)
	if err != nil {
		return nil, err
	}

	if network.Verbose {
		fmt.Println(fmt.Sprintf("\n[Harmony SDK]: %s - Generating a new transaction:\n\tReceiver address: %s\n\tAmount: %f\n\tNonce: %d\n\tGas limit: %d\n\tGas price: %f\n\tData length (bytes): %d\n",
			time.Now().Format(network.LoggingTimeFormat),
			toAddress,
			amount,
			nonce,
			calculatedGasLimit,
			gasPrice,
			len(inputData)),
		)
	}

	tx = transaction.NewTransaction(
		nonce,
		calculatedGasLimit,
		address.Parse(toAddress),
		amount.Mul(OneAsDec),
		gasPrice.Mul(NanoAsDec),
		[]byte(inputData),
	)

	return tx, nil
}

// SignTransaction - signs a transaction using a given keystore / account
func SignTransaction(keystore *keystore.KeyStore, account *accounts.Account, tx *transaction.Transaction, chainID *big.Int) (*types.Transaction, error) {
	signedTransaction, err := keystore.SignTx(*account, tx, chainID)
	if err != nil {
		return nil, err
	}

	return signedTransaction, nil
}

// AttachSigningData - attaches the signing data to the tx - necessary for e.g. propagating txs directly via p2p
func AttachSigningData(chainID *big.Int, tx *types.Transaction) (*types.Transaction, error) {
	signer := types.NewEIP155Signer(chainID)

	_, err := types.Sender(signer, tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// EncodeSignature - RLP encodes a given transaction signature as a hex signature
func EncodeSignature(tx interface{}) (*string, error) {
	enc, err := eth_rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}

	hexSignature := eth_hexutil.Encode(enc)
	signature := &hexSignature

	return signature, nil
}

// SendRawTransaction - sends a raw signed transaction via RPC
func SendRawTransaction(rpcClient *goSdkRPC.HTTPMessenger, signature *string) (interface{}, error) {

	rawTxBytes, err := eth_hexutil.Decode(*signature)
	tx := new(types.Transaction)
	eth_rlp.DecodeBytes(rawTxBytes, &tx)
	err = rpcClient.GetClient().SendTransaction(context.Background(), tx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("tx sent: %s \n", tx.Hash().Hex())
	return tx.Hash().Hex(), nil

}

// WaitForTxConfirmation - waits a given amount of seconds defined by timeout to try to receive a finalized transaction
func WaitForTxConfirmation(rpcClient *goSdkRPC.HTTPMessenger, node string, txType string, receiptHash string, timeout int) (map[string]interface{}, error) {
	//var failures []rpc.Failure

	if timeout > 0 {
		for {

			fmt.Println(fmt.Sprintf("wait for timeout %d", timeout))
			if timeout < 0 {
				return nil, nil
			}
			response, err := GetTransactionReceipt(rpcClient, receiptHash)
			if err != nil {
				return nil, err
			}

			if response != nil {
				return response, nil
			}

			time.Sleep(time.Second * 1)
			timeout = timeout - 1
		}
	}

	return nil, nil
}

func handleTransactionError(receiptHash string, failures []rpc.Failure) error {
	failure, failed := rpc.FailureOccurredForTransaction(failures, receiptHash)
	if failed {
		return errors.New(failure.ErrorMessage)
	}

	return nil
}

// CalculateGasLimit - calculates the proper gas limit for a given gas limit and input data
func CalculateGasLimit(gasLimit int64, inputData string, isValidatorCreation bool) (calculatedGasLimit uint64, err error) {
	// -1 means that the gas limit has not been specified by the user and that it should be automatically calculated based on the tx data
	if gasLimit == -1 {
		if len(inputData) > 0 {
			base64InputData, err := base64.StdEncoding.DecodeString(inputData)
			if err != nil {
				return 0, err
			}
			calculatedGasLimit, err = core.IntrinsicGasForStaking(base64InputData, isValidatorCreation)
			if err != nil {
				return 0, err
			}

			if calculatedGasLimit == 0 {
				return 0, errors.New("calculated gas limit is 0 - this shouldn't be possible")
			}
		} else {
			calculatedGasLimit = TxGasContractCreation
		}
	} else {
		calculatedGasLimit = uint64(gasLimit)
	}

	return calculatedGasLimit, nil
}

// BumpGasPrice - bumps the gas price by the required percentage, as defined by core.DefaultTxPoolConfig.PriceBump
func BumpGasPrice(gasPrice ethCommon.Dec) ethCommon.Dec {
	//return gasPrice.Add(ethCommon.NewDec(1).Quo(OneAsDec))
	return gasPrice.Mul(ethCommon.NewDec(100 + int64(core.DefaultTxPoolConfig.PriceBump)).Quo(ethCommon.NewDec(100)))
}

// GetTransactionReceipt - retrieves the transaction info/data for a transaction
func GetTransactionReceipt(rpcClient *goSdkRPC.HTTPMessenger, receiptHash interface{}) (map[string]interface{}, error) {

	receipt, err := rpcClient.GetClient().TransactionReceipt(context.Background(), ethCommon.HexToHash(receiptHash.(string)))

	if err != nil && err.Error() != "not found" {
		return nil, err
	}

	if receipt == nil {
		return nil, nil
	}

	result := make(map[string]interface{})
	result["transactionHash"] = receiptHash
	result["status"] = receipt.Status == 1
	var contractAddress string
	if !receipt.ContractAddress.IsEmpty() {
		contractAddress = address.ToBech32(receipt.ContractAddress)
	} else {
		contractAddress = ""
	}
	result["contractAddress"] = contractAddress

	return result, nil
}

// IsTransactionSuccessful - checks if a transaction is successful given a transaction response
func IsTransactionSuccessful(txResponse map[string]interface{}) (success bool) {
	txStatus, ok := txResponse["status"].(bool)

	if ok {
		success = txStatus
	}
	return success
}

// GenerateTxData - generates tx data based on a given byte size
func GenerateTxData(char string, byteSize int) string {
	buffer := new(bytes.Buffer)

	for i := 0; i < byteSize; i++ {
		buffer.Write([]byte(char))
	}

	return buffer.String()
}
