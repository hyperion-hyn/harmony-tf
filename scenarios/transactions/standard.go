package transactions

import (
	"fmt"
	"sync"
	"time"

	"github.com/harmony-one/harmony-tf/accounts"
	"github.com/harmony-one/harmony-tf/balances"
	"github.com/harmony-one/harmony-tf/config"
	"github.com/harmony-one/harmony-tf/funding"
	"github.com/harmony-one/harmony-tf/logger"
	"github.com/harmony-one/harmony-tf/testing"
	"github.com/harmony-one/harmony-tf/transactions"

	sdkAccounts "github.com/harmony-one/go-lib/accounts"
	sdkTxs "github.com/harmony-one/go-lib/transactions"
)

// StandardScenario - executes a standard/simple test case
func StandardScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)
	testCase.Executed = true
	if testCase.ReportError() {
		return
	}

	fundingAccountBalance := testing.RetrieveFundingAccountBalanceOrError(testCase)
	if testCase.Error != nil {
		return
	}

	fundingAmount, err := funding.CalculateFundingAmount(testCase.Parameters.Amount, fundingAccountBalance, testCase.Parameters.ReceiverCount)
	if err != nil {
		testCase.SetError(err)
		return
	}

	senderAccountName := accounts.GenerateTestCaseAccountName(testCase.Name, "Sender")
	logger.AccountLog(fmt.Sprintf("Generating a new sender account: %s", senderAccountName), testCase.Verbose)
	senderAccount, err := accounts.GenerateAccount(senderAccountName)
	if err != nil {
		testCase.SetError(err)
		return
	}

	logger.FundingLog(fmt.Sprintf("Funding sender account: %s, address: %s", senderAccount.Name, senderAccount.Address), testCase.Verbose)
	funding.PerformFundingTransaction(
		&config.Configuration.Funding.Account,
		testCase.Parameters.FromShardID,
		senderAccount.Address,
		testCase.Parameters.FromShardID,
		fundingAmount,
		-1,
		config.Configuration.Funding.Gas.Limit,
		config.Configuration.Funding.Gas.Price,
		config.Configuration.Funding.Timeout,
		config.Configuration.Funding.Retry.Attempts,
	)

	receiverAccountName := accounts.GenerateTestCaseAccountName(testCase.Name, "Receiver")
	logger.AccountLog(fmt.Sprintf("Generating a new receiver account: %s", receiverAccountName), testCase.Verbose)
	receiverAccount, err := accounts.GenerateAccount(receiverAccountName)

	senderStartingBalance, _ := balances.GetShardBalance(senderAccount.Address, testCase.Parameters.FromShardID)
	receiverStartingBalance, _ := balances.GetShardBalance(receiverAccount.Address, testCase.Parameters.ToShardID)
	txData := testCase.Parameters.GenerateTxData()

	logger.AccountLog(fmt.Sprintf("Using sender account %s, address: %s and receiver account %s, address : %s", senderAccount.Name, senderAccount.Address, receiverAccount.Name, receiverAccount.Address), testCase.Verbose)
	logger.BalanceLog(fmt.Sprintf("Sender account %s, address: %s has a starting balance of %f in shard %d before the test", senderAccount.Name, senderAccount.Address, senderStartingBalance, testCase.Parameters.FromShardID), testCase.Verbose)
	logger.BalanceLog(fmt.Sprintf("Receiver account %s, address: %s has a starting balance of %f in shard %d before the test", receiverAccount.Name, receiverAccount.Address, receiverStartingBalance, testCase.Parameters.ToShardID), testCase.Verbose)
	logger.TransactionLog(fmt.Sprintf("Sending transaction of %f token(s) from %s (shard %d) to %s (shard %d), tx data size: %d byte(s)", testCase.Parameters.Amount, senderAccount.Address, testCase.Parameters.FromShardID, receiverAccount.Address, testCase.Parameters.ToShardID, len(txData)), testCase.Verbose)
	logger.TransactionLog(fmt.Sprintf("Will wait up to %d seconds to let the transaction get finalized", testCase.Parameters.Timeout), testCase.Verbose)

	rawTx, err := transactions.SendTransaction(&senderAccount, testCase.Parameters.FromShardID, receiverAccount.Address, testCase.Parameters.ToShardID, testCase.Parameters.Amount, testCase.Parameters.Nonce, testCase.Parameters.Gas.Limit, testCase.Parameters.Gas.Price, txData, testCase.Parameters.Timeout)
	testCaseTx := sdkTxs.ToTransaction(senderAccount.Address, testCase.Parameters.FromShardID, receiverAccount.Address, testCase.Parameters.ToShardID, rawTx, err)
	testCase.Transactions = append(testCase.Transactions, testCaseTx)
	txResultColoring := logger.ResultColoring(testCaseTx.Success, true).Render(fmt.Sprintf("tx successful: %t", testCaseTx.Success))

	logger.TransactionLog(fmt.Sprintf("Sent %f token(s) from %s to %s - transaction hash: %s, %s", testCase.Parameters.Amount, senderAccount.Address, receiverAccount.Address, testCaseTx.TransactionHash, txResultColoring), testCase.Verbose)

	senderEndingBalance, _ := balances.GetShardBalance(senderAccount.Address, testCase.Parameters.FromShardID)

	if testCaseTx.Success && testCase.Parameters.FromShardID != testCase.Parameters.ToShardID {
		logger.TransactionLog(fmt.Sprintf("Because this is a cross shard transaction we need to wait an extra %d seconds to correctly receive the ending balance of the receiver account %s in shard %d", config.Configuration.Network.CrossShardTxWaitTime, receiverAccount.Address, testCase.Parameters.ToShardID), testCase.Verbose)
		time.Sleep(time.Duration(config.Configuration.Network.CrossShardTxWaitTime) * time.Second)
	}

	receiverEndingBalance, _ := balances.GetShardBalance(receiverAccount.Address, testCase.Parameters.ToShardID)
	expectedReceiverEndingBalance := receiverStartingBalance.Add(testCase.Parameters.Amount)
	testCase.Result = testCaseTx.Success && receiverEndingBalance.Equal(expectedReceiverEndingBalance)

	logger.BalanceLog(fmt.Sprintf("Sender address: %s has an ending balance of %f in shard %d after the test", senderAccount.Address, senderEndingBalance, testCase.Parameters.FromShardID), testCase.Verbose)
	logger.BalanceLog(fmt.Sprintf("Receiver address: %s has an ending balance of %f in shard %d after the test - expected balance is %f", receiverAccount.Address, receiverEndingBalance, testCase.Parameters.ToShardID, expectedReceiverEndingBalance), testCase.Verbose)
	logger.TeardownLog("Performing test teardown (returning funds and removing receiver account)\n", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)

	standardTeardown(testCase, senderAccount, receiverAccount)
	testing.Title(testCase, "footer", testCase.Verbose)
}

func standardTeardown(testCase *testing.TestCase, senderAccount sdkAccounts.Account, receiverAccount sdkAccounts.Account) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go testing.AsyncTeardown(&senderAccount, testCase.Parameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.Parameters.FromShardID, &waitGroup)
	go testing.AsyncTeardown(&receiverAccount, testCase.Parameters.ToShardID, config.Configuration.Funding.Account.Address, testCase.Parameters.FromShardID, &waitGroup)

	waitGroup.Wait()
}
