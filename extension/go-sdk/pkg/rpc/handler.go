package rpc

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Reply map[string]interface{}

type T interface {
	SendRPC(string, []interface{}) (Reply, error)
	GetClient() *ethclient.Client
}

type HTTPMessenger struct {
	node string
}

func (M *HTTPMessenger) SendRPC(meth string, params []interface{}) (Reply, error) {
	return Request(meth, M.node, params)
}

func (M *HTTPMessenger) GetClient() *ethclient.Client {
	client, err := ethclient.Dial(M.node)
	if err != nil {
		panic(fmt.Sprintf("Create eth client instance error %v", err))
	}
	return client
}

func NewHTTPHandler(node string) *HTTPMessenger {
	// TODO Sanity check the URL for HTTP
	return &HTTPMessenger{node}
}
