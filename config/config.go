package config

import (
	"time"

	"github.com/gookit/color"
	sdkAccounts "github.com/harmony-one/go-lib/accounts"
	sdkNetworkTypes "github.com/harmony-one/go-lib/network/types/network"
	sdkValidator "github.com/harmony-one/go-lib/staking/validator"
)

// Config - represents the general configuration
type Config struct {
	Framework  Framework `yaml:"framework"`
	Network    Network   `yaml:"network"`
	Account    Account   `yaml:"account"`
	Funding    Funding   `yaml:"funding"`
	Export     Export    `yaml:"export"`
	Configured bool
}

// Framework - represents common framework settings
type Framework struct {
	BasePath              string
	Identifier            string
	Version               string                  `yaml:"-"`
	Test                  string                  `yaml:"test"`
	Verbose               bool                    `yaml:"verbose"`
	MinimumRequiredMemory uint64                  `yaml:"minimum_required_memory"`
	SystemMemory          uint64                  `yaml:"-"` // In megabytes
	StartTime             time.Time               `yaml:"-"`
	EndTime               time.Time               `yaml:"-"`
	CurrentValidator      *sdkValidator.Validator `yaml:"-"`
	Styling               Styling                 `yaml:"-"`
}

// Styling - represents settings for styling the log output
type Styling struct {
	Header         *color.Style `yaml:"-"`
	TestCaseHeader *color.Style `yaml:"-"`
	Default        *color.Style `yaml:"-"`
	Account        *color.Style `yaml:"-"`
	Funding        *color.Style `yaml:"-"`
	Balance        *color.Style `yaml:"-"`
	Transaction    *color.Style `yaml:"-"`
	Staking        *color.Style `yaml:"-"`
	Teardown       *color.Style `yaml:"-"`
	Success        *color.Style `yaml:"-"`
	Warning        *color.Style `yaml:"-"`
	Error          *color.Style `yaml:"-"`
	Padding        string       `yaml:"-"`
}

// Network - represents the network settings group
type Network struct {
	Name                 string                  `yaml:"name"`
	Mode                 string                  `yaml:"mode"`
	Node                 string                  `yaml:"-"`
	Nodes                []string                `yaml:"-"`
	Endpoints            map[string][]string     `yaml:"endpoints"`
	Shards               int                     `yaml:"-"`
	CrossShardTxWaitTime uint32                  `yaml:"cross_shard_tx_wait_time"`
	StakingWaitTime      uint32                  `yaml:"staking_wait_time"`
	Gas                  sdkNetworkTypes.Gas     `yaml:"gas"`
	API                  sdkNetworkTypes.Network `yaml:"-"`
	Retry                Retry                   `yaml:"retry"`
	Balances             Balances                `yaml:"balances"`
}

// Account - represents the account settings group
type Account struct {
	Passphrase       string `yaml:"passphrase"`
	RemoveEmpty      bool   `yaml:"remove_empty"`
	UseAllInKeystore bool   `yaml:"use_all_in_keystore"`
}

// Funding - represents the funding settings group
type Funding struct {
	Account      sdkAccounts.Account `yaml:"account"`
	MinimumFunds float64             `yaml:"minimum_funds"`
	Timeout      int                 `yaml:"timeout"`
	Retry        Retry               `yaml:"retry"`
	Verbose      bool                `yaml:"verbose"`
	Shards       string              `yaml:"shards"`
	Gas          sdkNetworkTypes.Gas `yaml:"gas"`
}

// Retry - settings for RPC retries
type Retry struct {
	Attempts int `yaml:"attempts"`
	Wait     int `yaml:"wait"`
}

// Balances - settings for balance RPC calls
type Balances struct {
	Retry Retry `yaml:"retry"`
}

// Export - export settings
type Export struct {
	Path   string `yaml:"path"`
	Format string `yaml:"format"`
}

// Initialize - initializes basic framework settings
func (framework *Framework) Initialize() {
	if framework.MinimumRequiredMemory == 0 {
		framework.MinimumRequiredMemory = 8000
	}
}

// CanExecuteMemoryIntensiveTestCase - whether or not certain test cases can be executed due to heavy memory consumption
func (framework *Framework) CanExecuteMemoryIntensiveTestCase() bool {
	return framework.SystemMemory >= framework.MinimumRequiredMemory
}
