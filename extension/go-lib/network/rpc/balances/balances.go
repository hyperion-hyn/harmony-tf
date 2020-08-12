package balances

import (
	"context"
	ethCommon "github.com/ethereum/go-ethereum/common"
	goSDK_RPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
	"github.com/pkg/errors"
)

func GetBalance(messenger *goSDK_RPC.HTTPMessenger, address string) (ethCommon.Dec, error) {

	balance, err := messenger.GetClient().BalanceAt(context.Background(), ethCommon.HexToAddress(address), nil)
	if err != nil {
		return ethCommon.ZeroDec(), errors.Wrapf(err, "GetBalance")
	}

	return ethCommon.NewDecFromBigInt(balance), nil

}
