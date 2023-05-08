package cli

import (
	"fmt"
	"github.com/streamingfast/substreams/client"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	dauthAuthenticator "github.com/streamingfast/dauth/authenticator"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/dmetering"
	"github.com/streamingfast/dmetrics"
	"github.com/streamingfast/dstore"
	firehoseApp "github.com/streamingfast/firehose/app/firehose"
	"github.com/streamingfast/logging"
	substreamsService "github.com/streamingfast/substreams/service"
)

var metricset = dmetrics.NewSet()
var headBlockNumMetric = metricset.NewHeadBlockNumber("firehose")
var headTimeDriftmetric = metricset.NewHeadTimeDrift("firehose")

func init() {
	appLogger, _ := logging.PackageLogger("firehose", "github.com/pinax-network/firehose-antelope/firehose")

	launcher.RegisterApp(rootLog, &launcher.AppDef{
		ID:          "firehose",
		Title:       "Block Firehose",
		Description: "Provides on-demand filtered blocks, depends on common-merged-blocks-store-url and common-live-blocks-addr",
		RegisterFlags: func(cmd *cobra.Command) error {
			cmd.Flags().String("firehose-grpc-listen-addr", FirehoseGRPCServingAddr, "Address on which the firehose will listen")

			cmd.Flags().Bool("substreams-enabled", false, "Whether to enable substreams")
			cmd.Flags().Bool("substreams-partial-mode-enabled", false, "Whether to enable partial stores generation support on this instance (usually for internal deployments only)")
			cmd.Flags().Bool("substreams-request-stats-enabled", false, "Enables stats per request, like block rate. Should only be enabled in debugging instance not in production")
			cmd.Flags().String("substreams-state-store-url", "{sf-data-dir}/localdata", "where substreams state data are stored")
			cmd.Flags().Uint64("substreams-cache-save-interval", uint64(1_000), "Interval in blocks at which to save store snapshots and output caches")
			cmd.Flags().Uint64("substreams-max-fuel-per-block-module", uint64(5_000_000_000_000), "Hard limit for the number of instructions within the execution of a single wasmtime module for a single block")
			cmd.Flags().Int("substreams-parallel-subrequest-limit", 4, "number of parallel subrequests substream can make to synchronize its stores")
			cmd.Flags().String("substreams-client-endpoint", FirehoseGRPCServingAddr, "firehose endpoint for substreams client.")
			cmd.Flags().String("substreams-client-jwt", "", "JWT for substreams client authentication")
			cmd.Flags().Bool("substreams-client-insecure", false, "substreams client in insecure mode")
			cmd.Flags().Bool("substreams-client-plaintext", true, "substreams client in plaintext mode")
			cmd.Flags().Uint64("substreams-sub-request-parallel-jobs", 5, "substreams subrequest parallel jobs for the scheduler")
			cmd.Flags().Uint64("substreams-sub-request-block-range-size", 10000, "substreams subrequest block range size value for the scheduler")
			return nil
		},

		FactoryFunc: func(runtime *launcher.Runtime) (launcher.App, error) {
			sfDataDir := runtime.AbsDataDir

			blockstreamAddr := viper.GetString("common-live-blocks-addr")

			authenticator, err := dauthAuthenticator.New(viper.GetString("common-auth-plugin"))
			if err != nil {
				return nil, fmt.Errorf("unable to initialize dauth: %w", err)
			}

			metering, err := dmetering.New(viper.GetString("common-metering-plugin"))
			if err != nil {
				return nil, fmt.Errorf("unable to initialize dmetering: %w", err)
			}
			dmetering.SetDefaultMeter(metering)

			mergedBlocksStoreURL, oneBlocksStoreURL, forkedBlocksStoreURL, err := getCommonStoresURLs(runtime.AbsDataDir)
			if err != nil {
				return nil, err
			}

			var registerServiceExt firehoseApp.RegisterServiceExtensionFunc
			if viper.GetBool("substreams-enabled") {
				stateStore, err := dstore.NewStore(mustReplaceDataDir(sfDataDir, viper.GetString("substreams-state-store-url")), "", "", true)
				if err != nil {
					return nil, fmt.Errorf("setting up state store for data: %w", err)
				}

				opts := []substreamsService.Option{
					substreamsService.WithCacheSaveInterval(viper.GetUint64("substreams-cache-save-interval")),
					substreamsService.WithMaxWasmFuelPerBlockModule(viper.GetUint64("substreams-max-fuel-per-block-module")),
				}

				if viper.GetBool("substreams-request-stats-enabled") {
					opts = append(opts, substreamsService.WithRequestStats())
				}

				substreamsClientConfig := client.NewSubstreamsClientConfig(
					viper.GetString("substreams-client-endpoint"),
					os.ExpandEnv(viper.GetString("substreams-client-jwt")),
					viper.GetBool("substreams-client-insecure"),
					viper.GetBool("substreams-client-plaintext"),
				)

				if viper.GetBool("substreams-partial-mode-enabled") {
					tier2 := substreamsService.NewTier2(
						stateStore,
						"sf.ethereum.type.v2.Block",
						opts...,
					)

					registerServiceExt = tier2.Register
				} else {
					sss, err := substreamsService.NewTier1(
						stateStore,
						"sf.ethereum.type.v2.Block",
						viper.GetUint64("substreams-sub-request-parallel-jobs"),
						viper.GetUint64("substreams-sub-request-block-range-size"),
						substreamsClientConfig,
						opts...,
					)

					if err != nil {
						return nil, fmt.Errorf("creating substreams service: %w", err)
					}
					registerServiceExt = sss.Register
				}
			}

			return firehoseApp.New(appLogger, &firehoseApp.Config{
				OneBlocksStoreURL:       oneBlocksStoreURL,
				MergedBlocksStoreURL:    mergedBlocksStoreURL,
				ForkedBlocksStoreURL:    forkedBlocksStoreURL,
				BlockStreamAddr:         blockstreamAddr,
				GRPCListenAddr:          viper.GetString("firehose-grpc-listen-addr"),
				GRPCShutdownGracePeriod: 1 * time.Second,
			}, &firehoseApp.Modules{
				Authenticator:            authenticator,
				HeadTimeDriftMetric:      headTimeDriftmetric,
				HeadBlockNumberMetric:    headBlockNumMetric,
				RegisterServiceExtension: registerServiceExt,
			}), nil
		},
	})
}
