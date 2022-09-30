package cli

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dlauncher/launcher"
	mergerApp "github.com/streamingfast/merger/app/merger"
)

func init() {
	launcher.RegisterApp(rootLog, &launcher.AppDef{
		ID:          "merger",
		Title:       "Merger",
		Description: "Produces merged block files from single-block files",
		RegisterFlags: func(cmd *cobra.Command) error {
			cmd.Flags().Duration("merger-time-between-store-lookups", 1*time.Second, "Delay between source store polling (should be higher for remote storage)")
			cmd.Flags().Duration("merger-time-between-store-pruning", time.Minute, "Delay between source store pruning loops")
			cmd.Flags().Uint64("merger-prune-forked-blocks-after", 50000, "Number of blocks that must pass before we delete old forks (one-block-files lingering)")
			cmd.Flags().String("merger-grpc-listen-addr", MergerServingAddr, "Address to listen for incoming gRPC requests")
			return nil
		},
		// FIXME: Lots of config value construction is duplicated across InitFunc and FactoryFunc, how to streamline that
		//        and avoid the duplication? Note that this duplicate happens in many other apps, we might need to re-think our
		//        init flow and call init after the factory and giving it the instantiated app...
		InitFunc: func(runtime *launcher.Runtime) (err error) {
			sfDataDir := runtime.AbsDataDir

			if err = mkdirStorePathIfLocal(mustReplaceDataDir(sfDataDir, viper.GetString("common-merged-blocks-store-url"))); err != nil {
				return
			}

			if err = mkdirStorePathIfLocal(mustReplaceDataDir(sfDataDir, viper.GetString("common-one-block-store-url"))); err != nil {
				return
			}

			return nil
		},
		FactoryFunc: func(runtime *launcher.Runtime) (launcher.App, error) {
			sfDataDir := runtime.AbsDataDir

			return mergerApp.New(&mergerApp.Config{
				StorageOneBlockFilesPath:     MustReplaceDataDir(sfDataDir, viper.GetString("common-one-block-store-url")),
				StorageMergedBlocksFilesPath: MustReplaceDataDir(sfDataDir, viper.GetString("common-merged-blocks-store-url")),
				GRPCListenAddr:               viper.GetString("merger-grpc-listen-addr"),
				PruneForkedBlocksAfter:       viper.GetUint64("merger-prune-forked-blocks-after"),
				TimeBetweenPruning:           viper.GetDuration("merger-time-between-store-pruning"),
				TimeBetweenPolling:           viper.GetDuration("merger-time-between-store-lookups"),
			}), nil
		},
	})
}
