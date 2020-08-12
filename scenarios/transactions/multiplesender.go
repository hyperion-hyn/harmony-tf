package transactions

import (
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"sync"
	"time"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/balances"
	"github.com/hyperion-hyn/hyperion-tf/config"
	"github.com/hyperion-hyn/hyperion-tf/funding"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/testing"
	"github.com/hyperion-hyn/hyperion-tf/transactions"

	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkTxs "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
)

// Due to rate-limiting by RPC/explorer endpoints - don't run concurrently for now

// MultipleSenderScenario - runs a tests where multiple sender wallets are used to send to one respective new wallet
func MultipleSenderScenario(testCase *testing.TestCase) {
	testing.Title(testCase, "header", testCase.Verbose)

	if !config.Configuration.Framework.CanExecuteMemoryIntensiveTestCase() {
		testCase.ReportMemoryDismissal()
		return
	}

	testCase.Executed = true
	testCase.StartedAt = time.Now().UTC()

	if testCase.ErrorOccurred(nil) {
		return
	}

	_, _, err := funding.CalculateFundingDetails(testCase.Parameters.Amount, testCase.Parameters.SenderCount)
	if testCase.ErrorOccurred(err) {
		return
	}

	receiverAccountName := accounts.GenerateTestCaseAccountName(testCase.Name, "Receiver")
	receiverAccount, err := accounts.GenerateAccount(receiverAccountName)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate account %s", receiverAccountName)
		testCase.HandleError(err, &receiverAccount, msg)
		return
	}

	receiverStartingBalance, err := balances.GetBalance(receiverAccount.Address)
	if err != nil {
		msg := fmt.Sprintf("Failed to retrieve balance for account %s, address: %s", receiverAccount.Name, receiverAccount.Address)
		testCase.HandleError(err, &receiverAccount, msg)
		return
	}

	logger.BalanceLog(fmt.Sprintf("Receiver account %s (address: %s) has a starting balance of %f before the test", receiverAccount.Name, receiverAccount.Address, receiverStartingBalance), testCase.Verbose)

	nameTemplate := accounts.GenerateTestCaseAccountName(testCase.Name, "Sender_")
	senderAccounts, err := funding.GenerateAndFundAccounts(testCase.Parameters.SenderCount, nameTemplate, testCase.Parameters.Amount)
	if err != nil {
		msg := fmt.Sprintf("Failed to generate a total of %d sender accounts", testCase.Parameters.SenderCount)
		testCase.HandleError(err, nil, msg)
		return
	}

	executeMultiSenderTransactions(testCase, senderAccounts, receiverAccount)
	txsSuccessful := (testCase.SuccessfulTxCount == testCase.Parameters.SenderCount)

	logger.TransactionLog(fmt.Sprintf("A total of %d/%d transactions were successful", testCase.SuccessfulTxCount, testCase.Parameters.SenderCount), testCase.Verbose)

	if txsSuccessful {
		receiverEndingBalance, err := balances.GetNonZeroShardBalance(receiverAccount.Address)
		if testCase.ErrorOccurred(err) {
			return
		}
		decSenderCount := ethCommon.NewDec(testCase.Parameters.SenderCount)
		var expectedBalance ethCommon.Dec

		if testCase.Expected {
			expectedBalance = testCase.Parameters.Amount.Mul(decSenderCount)
		} else {
			expectedBalance = ethCommon.NewDec(0)
		}

		testCase.Result = (txsSuccessful && receiverEndingBalance.Equal(expectedBalance))

		logger.BalanceLog(fmt.Sprintf("Receiver account %s (address: %s) has an ending balance of %f after the test - expected balance: %f", receiverAccount.Name, receiverAccount.Address, receiverEndingBalance, expectedBalance), testCase.Verbose)
	} else {
		testCase.Result = false
	}

	logger.TeardownLog("Performing test teardown (returning funds and removing accounts)", testCase.Verbose)
	logger.ResultLog(testCase.Result, testCase.Expected, testCase.Verbose)

	multipleSendersTeardown(testCase, senderAccounts, receiverAccount)
	testing.Title(testCase, "footer", testCase.Verbose)

	testCase.FinishedAt = time.Now().UTC()
}

func executeMultiSenderTransactions(testCase *testing.TestCase, senderAccounts []sdkAccounts.Account, receiverAccount sdkAccounts.Account) {
	txs := make(chan sdkTxs.Transaction, testCase.Parameters.SenderCount)
	var waitGroup sync.WaitGroup

	for _, senderAccount := range senderAccounts {
		waitGroup.Add(1)
		go executeSenderTransaction(testCase, senderAccount, receiverAccount, txs, &waitGroup)
	}

	waitGroup.Wait()
	close(txs)

	testCase.SuccessfulTxCount = 0
	for tx := range txs {
		testCase.Transactions = append(testCase.Transactions, tx)
		if tx.Success {
			testCase.SuccessfulTxCount++
		}
	}
}

func executeSenderTransaction(testCase *testing.TestCase, senderAccount sdkAccounts.Account, receiverAccount sdkAccounts.Account, responses chan<- sdkTxs.Transaction, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	var testCaseTx sdkTxs.Transaction
	balanceRetrieved := true

	senderStartingBalance, err := balances.GetNonZeroShardBalance(senderAccount.Address)
	if err != nil {
		balanceRetrieved = false
	}

	if !senderStartingBalance.IsNil() && !senderStartingBalance.IsZero() {
		txData := testCase.Parameters.GenerateTxData()
		logger.BalanceLog(fmt.Sprintf("Sender account %s (address: %s) has a starting balance of %f before the test", senderAccount.Name, senderAccount.Address, senderStartingBalance), testCase.Verbose)
		logger.TransactionLog(fmt.Sprintf("Sending transaction of %f token(s) from %s to %s , tx data size: %d byte(s)", testCase.Parameters.Amount, senderAccount.Address, receiverAccount.Address, len(txData)), testCase.Verbose)
		logger.TransactionLog(fmt.Sprintf("Will wait up to %d seconds to let the transaction get finalized", testCase.Parameters.Timeout), testCase.Verbose)

		rawTx, err := transactions.SendTransaction(&senderAccount, receiverAccount.Address, testCase.Parameters.Amount, testCase.Parameters.Nonce, testCase.Parameters.Gas.Limit, testCase.Parameters.Gas.Price, txData, testCase.Parameters.Timeout)
		testCaseTx = sdkTxs.ToTransaction(senderAccount.Address, receiverAccount.Address, rawTx, err)
		if testCaseTx.Error != nil {
			logger.ErrorLog(fmt.Sprintf("Failed to send %f coins from %s  to %s  - error: %s", testCase.Parameters.Amount, senderAccount.Address, receiverAccount.Address, testCaseTx.Error.Error()), testCase.Verbose)
		} else {
			txResultColoring := logger.ResultColoring(testCaseTx.Success, true)
			logger.TransactionLog(fmt.Sprintf("Sent %f coins from %s to %s - transaction hash: %s, tx successful: %s", testCase.Parameters.Amount, senderAccount.Address, receiverAccount.Address, testCaseTx.TransactionHash, txResultColoring), testCase.Verbose)
		}
	} else {
		balanceRetrieved = false
	}

	if !balanceRetrieved {
		logger.FundingLog(fmt.Sprintf("Couldn't proceed with executing transaction since sender account %s hasn't been funded properly, balance is: %f", senderAccount.Address, senderStartingBalance), testCase.Verbose)

		testCaseTx = sdkTxs.Transaction{
			Success: false,
			Error:   fmt.Errorf("sender account %s wasn't funded properly, balance is: %f", senderAccount.Address, senderStartingBalance),
		}
	}

	responses <- testCaseTx
}

func multipleSendersTeardown(testCase *testing.TestCase, senderAccounts []sdkAccounts.Account, receiverAccount sdkAccounts.Account) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(1 + len(senderAccounts))

	for _, senderAccount := range senderAccounts {
		go testing.AsyncTeardown(&senderAccount, config.Configuration.Funding.Account.Address, &waitGroup)
	}
	go testing.AsyncTeardown(&receiverAccount, config.Configuration.Funding.Account.Address, &waitGroup)

	waitGroup.Wait()
}
