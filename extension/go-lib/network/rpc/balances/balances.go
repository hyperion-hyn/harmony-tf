package balances

import (
	"context"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	goSdkAddress "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/address"
	goSDK_RPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
	"github.com/pkg/errors"
)

func GetBalance(messenger *goSDK_RPC.HTTPMessenger, address string) (ethCommon.Dec, error) {

	balanceBigInt, err := messenger.GetClient().BalanceAt(context.Background(), goSdkAddress.Parse(address), nil)
	if err != nil {
		return ethCommon.ZeroDec(), errors.Wrapf(err, "GetBalance")
	}

	balance := ethCommon.NewDecFromBigInt(balanceBigInt)
	balance = balance.Quo(ethCommon.NewDec(params.Ether))

	return balance, nil

}
