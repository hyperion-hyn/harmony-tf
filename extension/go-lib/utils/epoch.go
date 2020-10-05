package utils

import (
	"context"
	"fmt"
	goSdkRPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
	"math/big"
	"time"
)

func WaitForEpoch(rpcClient *goSdkRPC.HTTPMessenger, skipEpoch int64) error {

	currentEpoch, err := CurrentEpoch(rpcClient)
	if err != nil {
		return err
	}

	targetEpoch := currentEpoch.Add(currentEpoch, big.NewInt(skipEpoch))

	fmt.Println(fmt.Sprintf("wait to epoch %d", targetEpoch))

	for {
		currentEpoch, err = CurrentEpoch(rpcClient)

		fmt.Println(fmt.Sprintf("current  epoch %d", currentEpoch))

		if err != nil {
			return err
		}
		if currentEpoch.Cmp(targetEpoch) >= 0 {
			break
		}
		time.Sleep(time.Duration(5) * time.Second)
	}
	return nil

}

func CurrentEpoch(rpcClient *goSdkRPC.HTTPMessenger) (*big.Int, error) {
	header, err := rpcClient.GetClient().HeaderByNumber(context.Background(), nil)

	if err != nil {
		return nil, err
	}
	return header.Epoch, nil
}
