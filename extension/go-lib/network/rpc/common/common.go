package common

import (
	goSDK_RPC "github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/rpc"
)

// NewRPCClient - resolve the RPC/HTTP Messenger to use for remote commands using a node and a shardID
func NewRPCClient(node string) (*goSDK_RPC.HTTPMessenger, error) {
	return goSDK_RPC.NewHTTPHandler(node), nil
}
