package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	dauthAuthenticator "github.com/streamingfast/dauth/authenticator"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/dmetering"
	"github.com/streamingfast/dmetrics"
	firehoseApp "github.com/streamingfast/firehose/app/firehose"
	"github.com/streamingfast/logging"
)

var metricset = dmetrics.NewSet()
var headBlockNumMetric = metricset.NewHeadBlockNumber("firehose")
var headTimeDriftmetric = metricset.NewHeadTimeDrift("firehose")

func init() {
	appLogger, _ := logging.PackageLogger("firehose", "github.com/streamingfast/firehose-acme/firehose")

	launcher.RegisterApp(rootLog, &launcher.AppDef{
		ID:          "firehose",
		Title:       "Block Firehose",
		Description: "Provides on-demand filtered blocks, depends on common-merged-blocks-store-url and common-relayer-addr",
		RegisterFlags: func(cmd *cobra.Command) error {
			cmd.Flags().String("firehose-grpc-listen-addr", FirehoseGRPCServingAddr, "Address on which the firehose will listen")
			cmd.Flags().Duration("firehose-real-time-tolerance", 1*time.Minute, "firehose will became alive if now - block time is smaller then tolerance")
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

			return firehoseApp.New(appLogger, &firehoseApp.Config{
				OneBlocksStoreURL:       MustReplaceDataDir(sfDataDir, viper.GetString("common-one-blocks-store-url")),
				MergedBlocksStoreURL:    MustReplaceDataDir(sfDataDir, viper.GetString("common-merged-blocks-store-url")),
				BlockStreamAddr:         viper.GetString("common-relayer-addr"),
				GRPCListenAddr:          viper.GetString("firehose-grpc-listen-addr"),
				GRPCShutdownGracePeriod: 1 * time.Second,
			}, &firehoseApp.Modules{
				Authenticator:         authenticator,
				HeadTimeDriftMetric:   headTimeDriftmetric,
				HeadBlockNumberMetric: headBlockNumMetric,
			}), nil
		},
	})
}
