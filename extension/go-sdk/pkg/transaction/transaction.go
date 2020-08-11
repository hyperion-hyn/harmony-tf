package transaction

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

type Transaction = types.Transaction

func NewTransaction(
	nonce, gasLimit uint64,
	to address.T,
	amount, gasPrice ethCommon.Dec,
	data []byte) *Transaction {
	return types.NewTransaction(nonce, to, amount.TruncateInt(), gasLimit, gasPrice.TruncateInt(), data[:])
}

// GetNextNonce returns the nonce on-chain (finalized transactions)
func GetNextNonce(addr string, messenger rpc.T) uint64 {
	transactionCountRPCReply, err :=
		messenger.SendRPC(rpc.Method.GetTransactionCount, []interface{}{address.Parse(addr), "latest"})

	if err != nil {
		return 0
	}

	transactionCount, _ := transactionCountRPCReply["result"].(string)
	n, _ := big.NewInt(0).SetString(transactionCount[2:], 16)
	return n.Uint64()
}

// GetNextPendingNonce returns the nonce from the tx-pool (un-finalized transactions)
func GetNextPendingNonce(addr string, messenger rpc.T) uint64 {
	transactionCountRPCReply, err :=
		messenger.SendRPC(rpc.Method.GetTransactionCount, []interface{}{address.Parse(addr), "pending"})

	if err != nil {
		return 0
	}

	transactionCount, _ := transactionCountRPCReply["result"].(string)
	n, _ := big.NewInt(0).SetString(transactionCount[2:], 16)
	return n.Uint64()
}

func IsValid(tx *Transaction) bool {
	return true
}
