package funding

import (
	"fmt"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"sync"

	"github.com/ethereum/go-ethereum/core"

	"github.com/hyperion-hyn/hyperion-tf/accounts"
	"github.com/hyperion-hyn/hyperion-tf/balances"
	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkNetworkNonce "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/rpc/nonces"
	sdkTransactions "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
	goSdkAccount "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/account"
	"github.com/hyperion-hyn/hyperion-tf/logger"
	"github.com/hyperion-hyn/hyperion-tf/transactions"
	"github.com/pkg/errors"
)

// SetupFundingAccount - sets up the initial funding account
func SetupFundingAccount(accs []sdkAccounts.Account) (err error) {
	if config.Configuration.Funding.Account.Address == "" {
		if sdkAccounts.DoesNamedAccountExist(config.Configuration.Funding.Account.Name) {
			if resolvedAddress := sdkAccounts.FindAccountAddressByName(config.Configuration.Funding.Account.Name); resolvedAddress != "" {
				config.Configuration.Funding.Account.Address = resolvedAddress
			}
		} else {
			config.Configuration.Funding.Account, err = accounts.GenerateAccount(config.Configuration.Funding.Account.Name)

			if err != nil {
				return err
			}
		}
	}

	config.Configuration.Funding.Account.Unlock()

	if len(accs) > 0 {
		logger.FundingLog(fmt.Sprintf("Proceeding to fund funding acccount %s / %s with a total of %d source accounts...", config.Configuration.Funding.Account.Name, config.Configuration.Funding.Account.Address, len(accs)), true)
		if err := FundFundingAccount(accs); err != nil {
			logger.ErrorLog(fmt.Sprintf("Proceeding to fund funding acccount %s / %s - error: %s", config.Configuration.Funding.Account.Name, config.Configuration.Funding.Account.Address, err.Error()), true)
		}
	}

	shardBalance, err := config.Configuration.Network.API.GetBalances(config.Configuration.Funding.Account.Address)
	if err != nil {
		return err
	}

	logger.BalanceLog(fmt.Sprintf("The current balance for the funding account %s / %s  is: %f", config.Configuration.Funding.Account.Name, config.Configuration.Funding.Account.Address, shardBalance), true)

	if shardBalance.IsNil() || shardBalance.IsZero() || shardBalance.IsNegative() {
		return fmt.Errorf(
			"Somehow the funding account %s, address: %s wasn't funded properly  - please make sure that you've added funded keystore files to keys/%s or that you've added funded private keys to keys/%s/private_keys.txt",
			config.Configuration.Funding.Account.Name,
			config.Configuration.Funding.Account.Address, config.Configuration.Network.Name,
			config.Configuration.Network.Name,
		)
	}

	if config.Configuration.Framework.Test == "all" && InsufficientBalance(shardBalance, config.Configuration.Funding.MinimumFunds) {
		logger.WarningLog(fmt.Sprintf("Funding account %s, address: %s wasn't funded properly , attempting to fund it", config.Configuration.Funding.Account.Name, config.Configuration.Funding.Account.Address), true)
		expectedBalance, err := fundFundingAccountInNonBeaconShard()
		if err != nil {
			return err
		}

		shardBalance, err = balances.GetExpectedShardBalance(config.Configuration.Funding.Account.Address, expectedBalance)
		if err != nil {
			return err
		}
		logger.BalanceLog(fmt.Sprintf("The current balance for the funding account %s / %s  is now: %f", config.Configuration.Funding.Account.Name, config.Configuration.Funding.Account.Address, shardBalance), true)
	}

	return nil
}

// FundFundingAccount - funds the funding account using the specified source accounts
func FundFundingAccount(accs []sdkAccounts.Account) error {
	var waitGroup sync.WaitGroup

	for _, account := range accs {
		if config.Configuration.Funding.Shards == "all" {
			FundFundingAccountInShard(account, &waitGroup)
		} else {
			//shard, err := strconv.ParseUint(config.Configuration.Funding.Shards, 10, 32)
			//if err != nil {
			//	return err
			//}

			FundFundingAccountInShard(account, &waitGroup)
		}
	}

	waitGroup.Wait()

	return nil
}

// FundFundingAccountInShard - fund the funding account in a given shard
func FundFundingAccountInShard(account sdkAccounts.Account, waitGroup *sync.WaitGroup) {
	availableShardBalance, err := balances.GetBalance(account.Address)

	if err == nil && !availableShardBalance.IsNil() && !availableShardBalance.IsZero() {
		amount := availableShardBalance.Sub(config.Configuration.Network.Gas.Cost)

		if !amount.IsNil() && !amount.IsZero() {
			waitGroup.Add(1)
			go AsyncPerformFundingTransaction(
				&account,
				config.Configuration.Funding.Account.Address,
				amount,
				-1,
				config.Configuration.Funding.Gas.Limit,
				config.Configuration.Funding.Gas.Price,
				config.Configuration.Funding.Timeout,
				config.Configuration.Funding.Retry.Attempts,
				waitGroup,
			)
		}
	} else if err != nil {
		fmt.Println(fmt.Sprintf("Failed to retrieve the shard balance for the address %s  - balance: %f, error: %s", account.Address, availableShardBalance, err.Error()))
	}
}

func fundFundingAccountInNonBeaconShard() (amount ethCommon.Dec, err error) {
	amount, err = CalculateFundingAmount(config.Configuration.Funding.MinimumFunds, 1)
	if err != nil {
		return ethCommon.NewDec(0), err
	}

	err = PerformFundingTransaction(
		&config.Configuration.Funding.Account,
		config.Configuration.Funding.Account.Address,
		amount,
		-1,
		config.Configuration.Funding.Gas.Limit,
		config.Configuration.Funding.Gas.Price,
		config.Configuration.Funding.Timeout,
		config.Configuration.Funding.Retry.Attempts,
	)
	if err != nil {
		return ethCommon.NewDec(0), err
	}

	return amount, nil
}

// GenerateAndFundAccounts - generate and fund a set of accounts
func GenerateAndFundAccounts(count int64, nameTemplate string, amount ethCommon.Dec) (accs []sdkAccounts.Account, err error) {
	rpcClient, err := config.Configuration.Network.API.RPCClient()
	if err != nil {
		return nil, errors.Wrapf(err, "RPC Client")
	}

	nonce := -1
	receivedNonce := sdkNetworkNonce.CurrentNonce(rpcClient, config.Configuration.Funding.Account.Address)
	if err != nil {
		return nil, errors.Wrapf(err, "Current Nonce")
	}
	nonce = int(receivedNonce)

	_, err = balances.GetBalance(config.Configuration.Funding.Account.Address)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("Shard Balance for %s ", config.Configuration.Funding.Account.Address))
	}

	amount, err = CalculateFundingAmount(amount, 1)
	if err != nil {
		return accs, errors.Wrapf(err, "Calculate Funding Amount")
	}

	var waitGroup sync.WaitGroup
	accountsChannel := make(chan sdkAccounts.Account, count)

	for i := int64(0); i < count; i++ {
		waitGroup.Add(1)
		go generateAndFundAccount(i, nameTemplate, amount, nonce, accountsChannel, &waitGroup)
		nonce++
	}

	waitGroup.Wait()
	close(accountsChannel)

	for acc := range accountsChannel {
		accs = append(accs, acc)
	}

	return accs, nil
}

func generateAndFundAccount(index int64, nameTemplate string, amount ethCommon.Dec, nonce int, accountsChannel chan<- sdkAccounts.Account, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	accountName := fmt.Sprintf("%s%d", nameTemplate, index)
	account, err := accounts.GenerateAccount(accountName)

	if err == nil {
		PerformFundingTransaction(
			&config.Configuration.Funding.Account,
			account.Address,
			amount,
			nonce,
			config.Configuration.Funding.Gas.Limit,
			config.Configuration.Funding.Gas.Price,
			config.Configuration.Funding.Timeout,
			config.Configuration.Funding.Retry.Attempts,
		)
		accountsChannel <- account
	}
}

// AsyncPerformFundingTransaction - performs an asynchronous call to PerformFundingTransaction and calls Done() on the waitGroup
func AsyncPerformFundingTransaction(account *sdkAccounts.Account, toAddress string, amount ethCommon.Dec, nonce int, gasLimit int64, gasPrice ethCommon.Dec, timeout int, attempts int, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	PerformFundingTransaction(account, toAddress, amount, nonce, gasLimit, gasPrice, timeout, attempts)
}

// FundAccounts - funds a given set of accounts in a given set of shards using a set of source accounts
func FundAccounts(sources []sdkAccounts.Account, count int, amount ethCommon.Dec, prefix string, gasLimit int64, gasPrice ethCommon.Dec, timeout int) (accounts []sdkAccounts.Account, err error) {
	for _, sourceAccount := range sources {
		for i := 0; i < count; i++ {
			account, err := fundAccount(&sourceAccount, amount, prefix, i, gasLimit, gasPrice, timeout)
			if err != nil {
				return accounts, err
			}

			if account.Name != "" && account.Address != "" {
				accounts = append(accounts, account)
			}
		}
	}

	return accounts, nil
}

func fundAccount(sourceAccount *sdkAccounts.Account, amount ethCommon.Dec, prefix string, index int, gasLimit int64, gasPrice ethCommon.Dec, timeout int) (account sdkAccounts.Account, err error) {
	accountName := fmt.Sprintf("%s_%d", prefix, index)

	// Remove the account just to make sure that we're starting using a clean slate
	goSdkAccount.RemoveAccount(accountName)

	account, err = accounts.GenerateAccount(accountName)
	if err != nil {
		return account, errors.Wrapf(err, "Generate Account")
	}

	err = PerformFundingTransaction(sourceAccount, account.Address, amount, -1, gasLimit, gasPrice, timeout, 10)

	if err != nil {
		return account, fmt.Errorf("failed to fund account %s  with amount %f using source account %s - error: %s", account.Address, amount, sourceAccount.Address, err.Error())
	}

	return account, nil
}

// PerformFundingTransaction - performs a funding transaction including automatic retries
func PerformFundingTransaction(account *sdkAccounts.Account, toAddress string, amount ethCommon.Dec, nonce int, gasLimit int64, gasPrice ethCommon.Dec, timeout int, attempts int) error {
	if amount.GT(ethCommon.NewDec(0)) {
		for {
			if attempts > 0 {
				logger.FundingLog(fmt.Sprintf("Attempting funding transaction from %s  to %s  of amount %f!", account.Address, toAddress, amount), config.Configuration.Funding.Verbose)

				rawTx, err := transactions.SendTransaction(account, toAddress, amount, nonce, gasLimit, gasPrice, "", config.Configuration.Funding.Timeout)

				if err != nil {
					if errors.Is(err, core.ErrUnderpriced) || errors.Is(err, core.ErrReplaceUnderpriced) || errors.Is(err, core.ErrIntrinsicGas) {
						gasPrice = sdkTransactions.BumpGasPrice(gasPrice)
						logger.ErrorLog(fmt.Sprintf("Failed to perform funding transaction from %s  to %s  of amount %f - error: %s", account.Address, toAddress, amount, err.Error()), config.Configuration.Funding.Verbose)
					} else if errors.Is(err, core.ErrInsufficientFunds) {
						return err
					}
				} else {
					success := sdkTransactions.IsTransactionSuccessful(rawTx)
					if success {
						logger.FundingLog(fmt.Sprintf("Successfully performed funding transaction (%s) from %s  to %s  of amount %f", rawTx["transactionHash"].(string), account.Address, toAddress, amount), config.Configuration.Funding.Verbose)
						break
					} else {
						gasPrice = sdkTransactions.BumpGasPrice(gasPrice)
						logger.FundingLog(fmt.Sprintf("Failed to perform funding transaction from %s  to %s of amount %f - retrying with new gas price: %f", account.Address, toAddress, amount, gasPrice), config.Configuration.Funding.Verbose)
					}
				}
			} else {
				return nil
			}

			attempts--
		}
	}

	return nil
}
