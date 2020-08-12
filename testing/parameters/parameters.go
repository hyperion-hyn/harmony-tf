package parameters

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	sdkAccounts "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/accounts"
	sdkNetworkTypes "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/network/types/network"
	sdkTxs "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/transactions"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-sdk/pkg/common"
	"github.com/pkg/errors"
)

// Parameters - represents the test case tx parameters
type Parameters struct {
	SenderCount   int64 `yaml:"sender_count"`
	Senders       []sdkAccounts.Account
	ReceiverCount int64 `yaml:"receiver_count"`
	Receivers     []sdkAccounts.Account
	Data          string              `yaml:"data"`
	DataSize      int                 `yaml:"data_size,omitempty"`
	RawAmount     string              `yaml:"amount"`
	Amount        ethCommon.Dec       `yaml:"-"`
	Gas           sdkNetworkTypes.Gas `yaml:"gas"`
	Nonce         int                 `yaml:"nonce"`
	Count         int                 `yaml:"count"`
	Timeout       int                 `yaml:"timeout"`
}

// Initialize - initializes and converts values for regular test case parameters
func (params *Parameters) Initialize() error {
	decAmount, err := common.NewDecFromString(params.RawAmount)
	if err != nil {
		return errors.Wrapf(err, "Parameters: Amount")
	}
	params.Amount = decAmount

	// This allocates unnecessary data on the heap
	/*if params.DataSize > 0 {
		params.Data = sdkTxs.GenerateTxData(params.Data, params.DataSize)
	}*/

	// Setup gas values
	if err := params.Gas.Initialize(); err != nil {
		return err
	}

	return nil
}

// GenerateTxData - generate tx data on the fly instead of allocating it
func (params *Parameters) GenerateTxData() string {
	if params.DataSize > 0 {
		return sdkTxs.GenerateTxData(params.Data, params.DataSize)
	}

	return ""
}
