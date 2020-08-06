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
	github.com/harmony-one/bls v0.0.6
	github.com/harmony-one/harmony v1.9.1-0.20200722170829-a354b93676e9
	github.com/karalabe/hid v1.0.0
	github.com/mackerelio/go-osstat v0.1.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/tyler-smith/go-bip39 v1.0.1-0.20181017060643-dbb3b84ba2ef
	github.com/valyala/fasthttp v1.2.0
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/ethereum/go-ethereum => ../go-ethereum
