package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

func NewStakingTransaction(
	nonce, gasLimit uint64, gasPrice *big.Int, txType types.TransactionType, payloadBytes []byte) (*types.Transaction, error) {
	coreTransaction := types.NewTransaction(nonce, common.BigToAddress(common.Big0), big.NewInt(0), gasLimit, gasPrice, payloadBytes)
	coreTransaction.SetType(txType)
	return coreTransaction, nil
}
