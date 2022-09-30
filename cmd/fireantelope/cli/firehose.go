package cli

import (
	"fmt"
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
	substreamsClient "github.com/streamingfast/substreams/client"
	substreamsService "github.com/streamingfast/substreams/service"
)

var metricset = dmetrics.NewSet()
var headBlockNumMetric = metricset.NewHeadBlockNumber("firehose")
var headTimeDriftmetric = metricset.NewHeadTimeDrift("firehose")

func init() {
	appLogger, _ := logging.PackageLogger("firehose", "github.com/streamingfast/firehose-acme/firehose")

	launcher.RegisterApp(rootLog, &launcher.AppDef{
		ID:          "firehose",
		Title:       "Block Firehose",
		Description: "Provides on-demand filtered blocks, depends on common-merged-blocks-store-url and common-live-blocks-addr",
		RegisterFlags: func(cmd *cobra.Command) error {
			cmd.Flags().String("firehose-grpc-listen-addr", FirehoseGRPCServingAddr, "Address on which the firehose will listen")

			cmd.Flags().Bool("substreams-enabled", false, "Whether to enable substreams")
			cmd.Flags().Bool("substreams-partial-mode-enabled", false, "Whether to enable partial stores generation support on this instance (usually for internal deployments only)")
			cmd.Flags().String("substreams-state-store-url", "{data-dir}/localdata", "where substreams state data are stored")
			cmd.Flags().Uint64("substreams-stores-save-interval", uint64(1_000), "Interval in blocks at which to save store snapshots")     // fixme
			cmd.Flags().Uint64("substreams-output-cache-save-interval", uint64(100), "Interval in blocks at which to save store snapshots") // fixme
			cmd.Flags().Int("substreams-parallel-subrequest-limit", 4, "number of parallel subrequests substream can make to synchronize its stores")
			cmd.Flags().String("substreams-client-endpoint", "", "Firehose endpoint for substreams client, if left empty, will default to this current local firehose.")
			cmd.Flags().String("substreams-client-jwt", "", "JWT for substreams client authentication")
			cmd.Flags().Bool("substreams-client-insecure", false, "Substreams client in insecure mode")
			cmd.Flags().Bool("substreams-client-plaintext", true, "Substreams client in plaintext mode")
			cmd.Flags().Int("substreams-sub-request-parallel-jobs", 5, "Substreams subrequest parallel jobs for the scheduler")
			cmd.Flags().Int("substreams-sub-request-block-range-size", 1000, "Substreams subrequest block range size value for the scheduler")
			return nil
		},

		FactoryFunc: func(runtime *launcher.Runtime) (launcher.App, error) {
			sfDataDir := runtime.AbsDataDir

			authenticator, err := dauthAuthenticator.New(viper.GetString("common-auth-plugin"))
			if err != nil {
				return nil, fmt.Errorf("unable to initialize dauth: %w", err)
			}

			metering, err := dmetering.New(viper.GetString("common-metering-plugin"))
			if err != nil {
				return nil, fmt.Errorf("unable to initialize dmetering: %w", err)
			}
			dmetering.SetDefaultMeter(metering)

			var registerServiceExt firehoseApp.RegisterServiceExtensionFunc
			if viper.GetBool("substreams-enabled") {
				stateStore, err := dstore.NewStore(MustReplaceDataDir(sfDataDir, viper.GetString("substreams-state-store-url")), "", "", true)
				if err != nil {
					return nil, fmt.Errorf("setting up state store for data: %w", err)
				}

				opts := []substreamsService.Option{
					substreamsService.WithStoresSaveInterval(viper.GetUint64("substreams-stores-save-interval")),
					substreamsService.WithOutCacheSaveInterval(viper.GetUint64("substreams-output-cache-save-interval")),
				}

				if viper.GetBool("substreams-partial-mode-enabled") {
					opts = append(opts, substreamsService.WithPartialMode())
				}

				clientEndpoint := viper.GetString("substreams-client-endpoint")
				if clientEndpoint == "" {
					clientEndpoint = viper.GetString("firehose-grpc-listen-addr")
				}

				clientConfig := substreamsClient.NewSubstreamsClientConfig(
					clientEndpoint,
					os.ExpandEnv(viper.GetString("substreams-client-jwt")),
					viper.GetBool("substreams-client-insecure"),
					viper.GetBool("substreams-client-plaintext"),
				)

				sss, err := substreamsService.New(
					stateStore,
					"sf.aptos.type.v1.Block",
					viper.GetInt("substreams-sub-request-parallel-jobs"),
					viper.GetInt("substreams-sub-request-block-range-size"),
					clientConfig,
					opts...,
				)
				if err != nil {
					return nil, fmt.Errorf("create substreams service: %w", err)
				}

				registerServiceExt = sss.Register
			}

			return firehoseApp.New(appLogger, &firehoseApp.Config{
				OneBlocksStoreURL:       MustReplaceDataDir(sfDataDir, viper.GetString("common-one-block-store-url")),
				MergedBlocksStoreURL:    MustReplaceDataDir(sfDataDir, viper.GetString("common-merged-blocks-store-url")),
				BlockStreamAddr:         viper.GetString("common-live-blocks-addr"),
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
