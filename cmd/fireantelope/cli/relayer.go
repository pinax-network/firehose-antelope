package cli

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dlauncher/launcher"
	relayerApp "github.com/streamingfast/relayer/app/relayer"
)

func init() {
	// Relayer
	launcher.RegisterApp(zlog, &launcher.AppDef{
		ID:          "relayer",
		Title:       "Relayer",
		Description: "Serves blocks as a stream, with a buffer",
		RegisterFlags: func(cmd *cobra.Command) error {
			cmd.Flags().String("relayer-grpc-listen-addr", RelayerServingAddr, "Address to listen for incoming gRPC requests")
			cmd.Flags().StringSlice("relayer-source", []string{ReaderGRPCAddr}, "List of Blockstream sources (reader-nodes) to connect to for live block feeds (repeat flag as needed)")
			cmd.Flags().Duration("relayer-max-source-latency", 999999*time.Hour, "Max latency tolerated to connect to a source. A performance optimization for when you have redundant sources and some may not have caught up")
			return nil
		},
		FactoryFunc: func(runtime *launcher.Runtime) (launcher.App, error) {
			_, oneBlocksStoreURL, _, err := getCommonStoresURLs(runtime.AbsDataDir)
			if err != nil {
				return nil, err
			}
			return relayerApp.New(&relayerApp.Config{
				SourceRequestBurst: 1,
				SourcesAddr:        viper.GetStringSlice("relayer-source"),
				OneBlocksURL:       oneBlocksStoreURL,
				GRPCListenAddr:     viper.GetString("relayer-grpc-listen-addr"),
				MaxSourceLatency:   viper.GetDuration("relayer-max-source-latency"),
			}), nil
		},
	})
}
