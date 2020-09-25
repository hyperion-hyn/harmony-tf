module github.com/hyperion-hyn/hyperion-tf

go 1.14

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/cosmos/cosmos-sdk v0.39.0
	github.com/deckarep/golang-set v1.7.1
	github.com/elliotchance/orderedmap v1.2.1
	github.com/ethereum/go-ethereum v1.8.27
	github.com/gookit/color v1.2.4
	github.com/hyperion-hyn/bls v0.0.6
	github.com/mackerelio/go-osstat v0.1.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/status-im/keycard-go v0.0.0-20190316090335-8537d3370df4
	github.com/tyler-smith/go-bip39 v1.0.1-0.20181017060643-dbb3b84ba2ef
	github.com/valyala/fasthttp v1.2.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/ethereum/go-ethereum => ../go-ethereum
