package config

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Args is a collection of global command arguments
var Args CommandArguments

// CommandArguments represents CLI command arguments
type CommandArguments struct {
	Network        string
	Mode           string
	Node           string
	Path           string
	Export         string
	ExportPath     string
	FundingAddress string
	MinimumFunds   float64
	Passphrase     string
	KeysPath       string
	TestTarget     string
	Verbose        bool
	VerboseGoSDK   bool
	PprofPort      int
}

var (
	// RootCommand - main entry point for Cobra commands
	RootCommand = &cobra.Command{
		Use:          "tests",
		Short:        "Regression tests",
		SilenceUsage: true,
		Long:         "Harmony regression testing tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	VersionWrap = fmt.Sprintf("github.com/SebastianJ (C) 2020. %v, version %s/%s-%s\n", path.Base(os.Args[0]), runtime.Version(), runtime.GOOS, runtime.GOARCH)
)

func init() {
	Args = CommandArguments{}
	RootCommand.PersistentFlags().StringVar(&Args.Network, "network", "stressnet", "--network <network>")
	RootCommand.PersistentFlags().StringVar(&Args.Mode, "mode", "api", "--mode <mode>")
	RootCommand.PersistentFlags().StringVar(&Args.Node, "node", "", "--node <node>")
	RootCommand.PersistentFlags().StringVar(&Args.Path, "path", ".", "<path>")
	RootCommand.PersistentFlags().StringVar(&Args.Export, "export", ".", "<path>")
	RootCommand.PersistentFlags().StringVar(&Args.ExportPath, "export-path", "./export", "<path>")
	RootCommand.PersistentFlags().StringVar(&Args.FundingAddress, "address", "", "--address <address>")
	RootCommand.PersistentFlags().Float64Var(&Args.MinimumFunds, "minimum-funds", 10.0, "--minimum-funds <funds>")
	RootCommand.PersistentFlags().StringVar(&Args.Passphrase, "passphrase", "", "--passphrase <passphrase>")
	RootCommand.PersistentFlags().StringVar(&Args.KeysPath, "keys", "", "--keys <path>")
	RootCommand.PersistentFlags().StringVar(&Args.TestTarget, "test", "", "--test <path>")
	RootCommand.PersistentFlags().BoolVar(&Args.Verbose, "verbose", false, "--verbose")
	RootCommand.PersistentFlags().BoolVar(&Args.VerboseGoSDK, "verbose-go-sdk", false, "--verbose-go-sdk")
	RootCommand.PersistentFlags().IntVar(&Args.PprofPort, "pprof-port", -1, "--pprof-port <port>")

	RootCommand.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stderr, VersionWrap)
			os.Exit(1)
			return nil
		},
	})
}

// Execute starts the actual app
func Execute() {
	RootCommand.SilenceErrors = true
	if err := RootCommand.Execute(); err != nil {
		fmt.Println(errors.Wrapf(err, "commit: %s, error", VersionWrap).Error())
		os.Exit(1)
	}
}
