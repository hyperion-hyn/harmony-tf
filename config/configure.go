package config

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/SebastianJ/harmony-tf/utils"
	"github.com/gookit/color"
	sdkNetwork "github.com/harmony-one/go-lib/network"
	sdkCommonTypes "github.com/harmony-one/go-lib/network/types/common"
	sdkNetworkTypes "github.com/harmony-one/go-lib/network/types/network"
	sdkNetworkUtils "github.com/harmony-one/go-lib/network/utils"
	goSdkCommon "github.com/harmony-one/go-sdk/pkg/common"
	"github.com/mackerelio/go-osstat/memory"
	"gopkg.in/yaml.v2"
)

// Configuration - the central configuration for the test suite tool
var (
	Configuration Config
)

// Configure - configures the test suite tool using a combination of the YAML config file as well as command arguments
func Configure(basePath string) (err error) {
	configPath := filepath.Join(basePath, "config.yml")
	if err = loadYamlConfig(configPath); err != nil {
		return err
	}

	if Configuration.Framework.BasePath == "" {
		Configuration.Framework.BasePath = basePath
	}

	ConfigureStylingConfig()

	if err = configureNetworkConfig(); err != nil {
		return err
	}

	if err = configureFrameworkConfig(); err != nil {
		return err
	}

	configureAccountConfig()

	if err = configureFundingConfig(); err != nil {
		return err
	}

	Configuration.Configured = true

	return nil
}

// ConfigureStylingConfig - configures the styling and color config
func ConfigureStylingConfig() {
	Configuration.Framework.Styling.Header = &color.Style{color.FgLightWhite, color.BgBlack, color.OpBold}
	Configuration.Framework.Styling.TestCaseHeader = &color.Style{color.FgLightWhite, color.BgGray, color.OpBold}
	Configuration.Framework.Styling.Default = &color.Style{color.OpReset}
	Configuration.Framework.Styling.Account = &color.Style{color.FgCyan, color.OpBold}
	Configuration.Framework.Styling.Funding = &color.Style{color.FgMagenta, color.OpBold}
	Configuration.Framework.Styling.Balance = &color.Style{color.FgLightBlue, color.OpBold}
	Configuration.Framework.Styling.Transaction = &color.Style{color.FgYellow, color.OpBold}
	Configuration.Framework.Styling.Staking = &color.Style{color.FgLightGreen, color.OpBold}
	Configuration.Framework.Styling.Teardown = &color.Style{color.FgGray, color.OpBold}
	Configuration.Framework.Styling.Success = &color.Style{color.FgLightWhite, color.BgGreen}
	Configuration.Framework.Styling.Warning = &color.Style{color.FgLightWhite, color.BgYellow}
	Configuration.Framework.Styling.Error = &color.Style{color.FgLightWhite, color.BgRed}
	Configuration.Framework.Styling.Padding = strings.Repeat("\t", 10)
}

func configureNetworkConfig() error {
	if Args.Network != "" && Args.Network != Configuration.Network.Name {
		Configuration.Network.Name = Args.Network
	}

	Configuration.Network.Name = sdkNetworkUtils.NormalizedNetworkName(Configuration.Network.Name)
	if Configuration.Network.Name == "" {
		return errors.New("you need to specify a valid network name to use! Valid options: localnet, devnet, testnet, staking, stressnet, mainnet")
	}

	for networkType, nodes := range Configuration.Network.Endpoints {
		if networkType == Configuration.Network.Name {
			Configuration.Network.Nodes = nodes
			break
		}
	}

	Configuration.Network.Mode = strings.ToLower(Configuration.Network.Mode)
	mode := strings.ToLower(Args.Mode)
	if mode != "" && mode != Configuration.Network.Mode {
		Configuration.Network.Mode = mode
	}

	node := sdkNetworkUtils.ResolveStartingNode(Configuration.Network.Name, Configuration.Network.Mode, 0, Configuration.Network.Nodes)
	shards, shardingStructure, err := sdkNetworkTypes.GenerateShardSetup(node, Configuration.Network.Name, Configuration.Network.Mode, Configuration.Network.Nodes)
	if err != nil {
		return fmt.Errorf("failed to generate network & shard setup for network %s using node %s - error: %s", Configuration.Network.Name, node, err.Error())
	}

	if Configuration.Network.Mode == "api" {
		Configuration.Network.Nodes = []string{}
		for _, shard := range shards {
			Configuration.Network.Nodes = append(Configuration.Network.Nodes, shard.Node)
		}
	}

	Configuration.Network.API = sdkNetworkTypes.Network{
		Name:              Configuration.Network.Name,
		Mode:              Configuration.Network.Mode,
		Shards:            shards,
		ShardingStructure: shardingStructure,
		Retry: sdkCommonTypes.Retry{
			Attempts: Configuration.Network.Retry.Attempts,
			Wait:     Configuration.Network.Retry.Wait,
		},
	}

	Configuration.Network.API.Initialize()

	if Configuration.Network.API.ChainID == nil {
		return errors.New("chain id must be set - please check that you are using correct network settings")
	}

	Configuration.Network.Shards = len(shardingStructure)

	if err := Configuration.Network.Gas.Initialize(); err != nil {
		return err
	}

	return nil
}

func configureFrameworkConfig() error {
	Configuration.Framework.Identifier = "HarmonyTF"
	Configuration.Framework.Version = "0.0.1"

	Configuration.Framework.Verbose = Args.Verbose
	// Set the verbosity level of harmony-sdk
	sdkNetwork.Verbose = Configuration.Framework.Verbose

	// Set the verbosity level of go-sdk
	goSdkCommon.DebugRPC = Args.VerboseGoSDK

	totalMemory, err := availableTotalMemory()
	if err != nil {
		return err
	}
	Configuration.Framework.SystemMemory = totalMemory

	Configuration.Framework.StartTime = time.Now()

	testTarget := strings.ToLower(Args.TestTarget)
	if testTarget != "" && testTarget != Configuration.Framework.Test {
		Configuration.Framework.Test = testTarget
	}
	if Configuration.Framework.Test == "" {
		Configuration.Framework.Test = "all"
	}

	testType := strings.ToLower(Configuration.Framework.Test)
	switch testType {
	case "txs", "transactions":
		Configuration.Framework.Test = "transactions"
	case "staking", "validator", "stake":
		Configuration.Framework.Test = "staking"
	default:
		Configuration.Framework.Test = testType
	}

	return nil
}

func configureAccountConfig() {
	if Args.Passphrase != "" && Args.Passphrase != Configuration.Account.Passphrase {
		Configuration.Account.Passphrase = Args.Passphrase
	}
}

func configureFundingConfig() error {
	Configuration.Funding.Account.Name = fmt.Sprintf("%s_%s_%s", Configuration.Framework.Identifier, strings.Title(Configuration.Network.Name), Configuration.Funding.Account.Name)

	if Args.FundingAddress != "" && Args.FundingAddress != Configuration.Funding.Account.Address {
		Configuration.Funding.Account.Address = Args.FundingAddress
	}

	if Configuration.Framework.Test == "staking" {
		Configuration.Funding.Shards = "0"
	} else {
		Configuration.Funding.Shards = strings.ToLower(Configuration.Funding.Shards)
		if Configuration.Funding.Shards == "" {
			Configuration.Funding.Shards = "all"
		}
	}

	Configuration.Funding.Timeout = utils.NetworkTimeoutAdjustment(Configuration.Network.Name, Configuration.Funding.Timeout)

	if Args.MinimumFunds != 0.0 && Args.MinimumFunds != Configuration.Funding.MinimumFunds {
		Configuration.Funding.MinimumFunds = Args.MinimumFunds
	}

	if err := Configuration.Funding.Gas.Initialize(); err != nil {
		return err
	}

	return nil
}

func loadYamlConfig(path string) error {
	Configuration = Config{}
	yamlData, err := utils.ReadFileToString(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(yamlData), &Configuration)
	if err != nil {
		return err
	}

	return nil
}

func availableTotalMemory() (uint64, error) {
	memory, err := memory.Get()
	if err != nil {
		return 0, err
	}

	totalMemoryMb := uint64(math.RoundToEven(float64(memory.Total) / float64(1024*1024)))

	return totalMemoryMb, nil
}
