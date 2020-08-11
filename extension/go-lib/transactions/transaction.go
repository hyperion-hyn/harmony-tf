package transactions

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Transaction - represents an executed test case transaction
type Transaction struct {
	FromAddress     string
	ToAddress       string
	Data            string
	Amount          ethCommon.Dec
	GasPrice        int64
	Timeout         int
	TransactionHash string
	Success         bool
	Response        map[string]interface{}
	Error           error
}

type StakeMsgFulfiller func() (types.TransactionType, interface{})

// ToTransaction - converts a raw tx response map to a typed Transaction type
func ToTransaction(fromAddress string, toAddress string, rawTx map[string]interface{}, err error) Transaction {
	if err != nil {
		return Transaction{Error: err}
	}

	var tx Transaction

	txHash := rawTx["transactionHash"].(string)

	if txHash != "" {
		success := IsTransactionSuccessful(rawTx)

		tx = Transaction{
			FromAddress:     fromAddress,
			ToAddress:       toAddress,
			TransactionHash: txHash,
			Success:         success,
			Response:        rawTx,
		}
	}

	return tx
}
