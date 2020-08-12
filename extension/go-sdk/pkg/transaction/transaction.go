package transaction

import (
	"context"
	ethCommon "github.com/ethereum/go-ethereum/common"
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
	nonce, err := messenger.GetClient().NonceAt(context.Background(), address.Parse(addr), nil)
	if err != nil {
		return 0
	}
	return nonce
}
