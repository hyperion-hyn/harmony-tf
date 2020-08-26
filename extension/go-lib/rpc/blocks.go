package rpc

import (
	"time"

	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/utils"
)

// BlockWrapper - wrapper for the GetBlockByNumber RPC method
type BlockWrapper struct {
	ID      string    `json:"id" yaml:"id"`
	JSONRPC string    `json:"jsonrpc" yaml:"jsonrpc"`
	Result  BlockInfo `json:"result" yaml:"result"`
	Error   RPCError  `json:"error,omitempty" yaml:"error,omitempty"`
}

// BlockInfo - block info
type BlockInfo struct {
	BlockNumber  uint64    `json:"-" yaml:"-"`
	Difficulty   int       `json:"difficulty,omitempty" yaml:"difficulty,omitempty"`
	ExtraData    string    `json:"extraData,omitempty" yaml:"extraData,omitempty"`
	Hash         string    `json:"hash,omitempty" yaml:"hash,omitempty"`
	Nonce        uint32    `json:"nonce,omitempty" yaml:"nonce,omitempty"`
	RawTimestamp string    `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
	Timestamp    time.Time `json:"-" yaml:"-"`
}

// RPCGenericSingleHexResponse - wrapper for RPC calls returning a single result in a hex format
type RPCGenericSingleHexResponse struct {
	ID      string `json:"id" yaml:"id"`
	JSONRPC string `json:"jsonrpc" yaml:"jsonrpc"`
	Result  string `json:"result" yaml:"result"`
}

// Initialize - initialize and convert values for a given BlockInfo struct
func (blockInfo *BlockInfo) Initialize() error {
	if blockInfo.RawTimestamp != "" {
		unixTime, err := utils.HexToDecimal(blockInfo.RawTimestamp)
		if err != nil {
			return err
		}

		blockInfo.Timestamp = time.Unix(int64(unixTime), 0).UTC()
	}

	return nil
}
