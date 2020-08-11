package testing

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"sync"

	"github.com/hyperion-hyn/hyperion-tf/balances"
	"github.com/hyperion-hyn/hyperion-tf/config"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	goSdkAccount "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/account"
	"github.com/hyperion-hyn/hyperion-tf/transactions"
)

// Teardown - return any sent tokens (minus a gas cost) and remove the account from the keystore
func Teardown(account *sdkAccounts.Account, fromShardID uint32, toAddress string, toShardID uint32) {
	amount, err := balances.GetShardBalance(account.Address, fromShardID)

	if err == nil && !amount.IsNil() {
		if amount.GT(ethCommon.NewDec(0)) {
			amount = amount.Sub(config.Configuration.Funding.Gas.Cost)
		}

		if amount.GT(ethCommon.NewDec(0)) {
			transactions.SendTransaction(account, fromShardID, toAddress, toShardID, amount, -1, config.Configuration.Funding.Gas.Limit, config.Configuration.Funding.Gas.Price, "", 0)
		}
	}

	goSdkAccount.RemoveAccount(account.Name)
}

// AsyncTeardown - return any sent tokens (minus a gas cost) and remove the account from the keystore
func AsyncTeardown(account *sdkAccounts.Account, fromShardID uint32, toAddress string, toShardID uint32, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	Teardown(account, fromShardID, toAddress, toShardID)
}
