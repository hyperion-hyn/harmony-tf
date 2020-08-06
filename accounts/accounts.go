package accounts

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	goSdkAccount "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/account"
)

// GenerateTestCaseAccountName - generate a test case prefixed account name
func GenerateTestCaseAccountName(testCaseID string, name string) string {
	return GenerateAccountName(fmt.Sprintf("TestCase_%s_%s", testCaseID, name))
}

// GenerateAccountName - generate a prefixed account name
func GenerateAccountName(name string) string {
	return fmt.Sprintf("%s_%s_%s",
		config.Configuration.Framework.Identifier,
		strings.Title(config.Configuration.Network.Name),
		name,
	)
}

// GenerateAccount - wrapper around sdkAccounts.GenerateAccount
func GenerateAccount(name string) (sdkAccounts.Account, error) {
	return PerformGenerateAccount(name, 3)
}

// PerformGenerateAccount - wrapper around sdkAccounts.GenerateAccount
func PerformGenerateAccount(name string, attempts int) (sdkAccounts.Account, error) {
	account, err := sdkAccounts.GenerateAccount(name, config.Configuration.Account.Passphrase)

	for {
		if (err != nil || account.Name == "" || account.Address == "") && attempts > 0 {
			goSdkAccount.RemoveAccount(name)
			account, err = sdkAccounts.GenerateAccount(name, config.Configuration.Account.Passphrase)
			attempts--
		} else {
			break
		}
	}

	return account, err
}

// ImportPrivateKeyAccount - wrapper around sdkAccounts.ImportPrivateKeyAccount
func ImportPrivateKeyAccount(privateKey string, keyName string, address string) (sdkAccounts.Account, error) {
	return sdkAccounts.ImportPrivateKeyAccount(privateKey, keyName, address, config.Configuration.Account.Passphrase)
}

// ImportKeystoreAccount - wrapper around sdkAccounts.ImportKeystoreAccount
func ImportKeystoreAccount(keyFile string, keyName string, address string) (sdkAccounts.Account, error) {
	return sdkAccounts.ImportKeystoreAccount(keyFile, keyName, address, config.Configuration.Account.Passphrase)
}

// GenerateMultipleAccounts - generates multiple typed accounts
func GenerateMultipleAccounts(nameTemplate string, count int64) (accs []sdkAccounts.Account) {
	for i := int64(0); i < count; i++ {
		name := fmt.Sprintf("%s%d", nameTemplate, i)
		acc, err := GenerateAccount(name)

		if err == nil && acc.Address != "" {
			accs = append(accs, acc)
		}
	}

	return accs
}

// AsyncGenerateMultipleAccounts - asynchronously generates multiple typed accounts
func AsyncGenerateMultipleAccounts(nameTemplate string, count int64) (accs []sdkAccounts.Account) {
	accountsChannel := make(chan sdkAccounts.Account, count)
	var waitGroup sync.WaitGroup

	for i := int64(0); i < count; i++ {
		name := fmt.Sprintf("%s%d", nameTemplate, i)
		waitGroup.Add(1)
		go func(name string, accountsChannel chan<- sdkAccounts.Account, waitGroup *sync.WaitGroup) {
			defer waitGroup.Done()
			acc, err := GenerateAccount(name)

			if err == nil && acc.Address != "" {
				accountsChannel <- acc
			}
		}(name, accountsChannel, &waitGroup)
	}

	waitGroup.Wait()
	close(accountsChannel)

	for acc := range accountsChannel {
		accs = append(accs, acc)
	}

	return accs
}
