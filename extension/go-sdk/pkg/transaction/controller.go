package transaction

import (
	"errors"
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/ledger"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

var (
	nanoAsDec = ethCommon.NewDec(params.GWei)
	oneAsDec  = ethCommon.NewDec(params.Ether)

	// ErrBadTransactionParam is returned when invalid params are given to the
	// controller upon execution of a transaction.
	ErrBadTransactionParam = errors.New("transaction has bad parameters")
)

type p []interface{}

type transactionForRPC struct {
	params      map[string]interface{}
	transaction *Transaction
	// Hex encoded
	signature       *string
	transactionHash *string
	receipt         rpc.Reply
}

type sender struct {
	ks      *keystore.KeyStore
	account *accounts.Account
}

// Controller drives the transaction signing process
type Controller struct {
	executionError    error
	transactionErrors Errors
	messenger         rpc.T
	sender            sender
	transactionForRPC transactionForRPC
	chain             common.ChainID
	Behavior          behavior
}

type behavior struct {
	DryRun               bool
	SigningImpl          SignerImpl
	ConfirmationWaitTime uint32
}

// TransactionToJSON dumps JSON representation
func (C *Controller) TransactionToJSON(pretty bool) string {
	r, _ := C.transactionForRPC.transaction.MarshalJSON()
	if pretty {
		return common.JSONPrettyFormat(string(r))
	}
	return string(r)
}

// RawTransaction dumps the signature as string
func (C *Controller) RawTransaction() string {
	return *C.transactionForRPC.signature
}

func (C *Controller) TransactionHash() *string {
	return C.transactionForRPC.transactionHash
}

func (C *Controller) Receipt() rpc.Reply {
	return C.transactionForRPC.receipt
}

func (C *Controller) TransactionErrors() Errors {
	return C.transactionErrors
}

func (C *Controller) setShardIDs(fromShard, toShard uint32) {
	if C.executionError != nil {
		return
	}
	C.transactionForRPC.params["from-shard"] = fromShard
	C.transactionForRPC.params["to-shard"] = toShard
}

func (C *Controller) setIntrinsicGas(gasLimit uint64) {
	if C.executionError != nil {
		return
	}
	C.transactionForRPC.params["gas-limit"] = gasLimit
}

func (C *Controller) setGasPrice(gasPrice ethCommon.Dec) {
	if C.executionError != nil {
		return
	}
	if gasPrice.IsNegative() {
		C.executionError = ErrBadTransactionParam
		errorMsg := fmt.Sprintf(
			"can't set negative gas price: %d", gasPrice,
		)
		C.transactionErrors = append(C.transactionErrors, &Error{
			ErrMessage:           &errorMsg,
			TimestampOfRejection: time.Now().Unix(),
		})
		return
	}
	C.transactionForRPC.params["gas-price"] = gasPrice.Mul(nanoAsDec)
}

func (C *Controller) setAmount(amount ethCommon.Dec) {
	if C.executionError != nil {
		return
	}
	if amount.IsNegative() {
		C.executionError = ErrBadTransactionParam
		errorMsg := fmt.Sprintf(
			"can't set negative amount: %d", amount,
		)
		C.transactionErrors = append(C.transactionErrors, &Error{
			ErrMessage:           &errorMsg,
			TimestampOfRejection: time.Now().Unix(),
		})
		return
	}
	balanceRPCReply, err := C.messenger.SendRPC(
		rpc.Method.GetBalance,
		p{address.ToBech32(C.sender.account.Address), "latest"},
	)
	if err != nil {
		C.executionError = err
		return
	}
	currentBalance, _ := balanceRPCReply["result"].(string)
	bal, _ := new(big.Int).SetString(currentBalance[2:], 16)
	balance := ethCommon.NewDecFromBigInt(bal)
	gasAsDec := C.transactionForRPC.params["gas-price"].(ethCommon.Dec)
	gasAsDec = gasAsDec.Mul(ethCommon.NewDec(int64(C.transactionForRPC.params["gas-limit"].(uint64))))
	amountInAtto := amount.Mul(oneAsDec)
	total := amountInAtto.Add(gasAsDec)

	if total.GT(balance) {
		balanceInOne := balance.Quo(oneAsDec)
		C.executionError = ErrBadTransactionParam
		errorMsg := fmt.Sprintf(
			"insufficient balance of %s in shard %d for the requested transfer of %s",
			balanceInOne.String(), C.transactionForRPC.params["from-shard"].(uint32), amount.String(),
		)
		C.transactionErrors = append(C.transactionErrors, &Error{
			ErrMessage:           &errorMsg,
			TimestampOfRejection: time.Now().Unix(),
		})
		return
	}
	C.transactionForRPC.params["transfer-amount"] = amountInAtto
}

func (C *Controller) setReceiver(receiver string) {
	C.transactionForRPC.params["receiver"] = address.Parse(receiver)
}

func (C *Controller) signAndPrepareTxEncodedForSending() {
	if C.executionError != nil {
		return
	}
	signedTransaction, err :=
		C.sender.ks.SignTx(*C.sender.account, C.transactionForRPC.transaction, C.chain.Value)
	if err != nil {
		C.executionError = err
		return
	}
	C.transactionForRPC.transaction = signedTransaction
	enc, _ := rlp.EncodeToBytes(signedTransaction)
	hexSignature := hexutil.Encode(enc)
	C.transactionForRPC.signature = &hexSignature
	if common.DebugTransaction {
		r, _ := signedTransaction.MarshalJSON()
		//fmt.Println("Signed with ChainID:", C.transactionForRPC.transaction.ChainID())
		fmt.Println(common.JSONPrettyFormat(string(r)))
	}
}

func (C *Controller) hardwareSignAndPrepareTxEncodedForSending() {
	if C.executionError != nil {
		return
	}
	enc, signerAddr, err := ledger.SignTx(C.transactionForRPC.transaction, C.chain.Value)
	if err != nil {
		C.executionError = err
		return
	}
	if strings.Compare(signerAddr, address.ToBech32(C.sender.account.Address)) != 0 {
		C.executionError = ErrBadTransactionParam
		errorMsg := "signature verification failed : sender address doesn't match with ledger hardware addresss"
		C.transactionErrors = append(C.transactionErrors, &Error{
			ErrMessage:           &errorMsg,
			TimestampOfRejection: time.Now().Unix(),
		})
		return
	}
	hexSignature := hexutil.Encode(enc)
	C.transactionForRPC.signature = &hexSignature
}

func (C *Controller) sendSignedTx() {
	if C.executionError != nil || C.Behavior.DryRun {
		return
	}
	reply, err := C.messenger.SendRPC(rpc.Method.SendRawTransaction, p{C.transactionForRPC.signature})
	if err != nil {
		C.executionError = err
		return
	}
	r, _ := reply["result"].(string)
	C.transactionForRPC.transactionHash = &r
}

func (C *Controller) txConfirmation() {
	if C.executionError != nil || C.Behavior.DryRun {
		return
	}
	if C.Behavior.ConfirmationWaitTime > 0 {
		txHash := *C.TransactionHash()
		start := int(C.Behavior.ConfirmationWaitTime)
		for {
			r, _ := C.messenger.SendRPC(rpc.Method.GetTransactionReceipt, p{txHash})
			if r["result"] != nil {
				C.transactionForRPC.receipt = r
				return
			}
			transactionErrors, err := GetError(txHash, C.messenger)
			if err != nil {
				errMsg := fmt.Sprintf(err.Error())
				C.transactionErrors = append(C.transactionErrors, &Error{
					TxHashID:             &txHash,
					ErrMessage:           &errMsg,
					TimestampOfRejection: time.Now().Unix(),
				})
			}
			C.transactionErrors = append(C.transactionErrors, transactionErrors...)
			if len(transactionErrors) > 0 {
				C.executionError = fmt.Errorf("error found for transaction hash: %s", txHash)
				return
			}
			if start < 0 {
				C.executionError = fmt.Errorf("could not confirm transaction after %d seconds", C.Behavior.ConfirmationWaitTime)
				return
			}
			time.Sleep(time.Second)
			start -= 1
		}
	}
}

// TODO: add logic to create staking transactions in the SDK.
