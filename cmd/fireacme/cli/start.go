// Copyright 2021 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dlauncher/launcher"
	"go.uber.org/zap"
)

var StartCmd = &cobra.Command{Use: "start", Short: "Starts `fireacme` services all at once", RunE: firehoseCmdStartE, Args: cobra.ArbitraryArgs}

func init() {
	RootCmd.AddCommand(StartCmd)
}

func firehoseCmdStartE(cmd *cobra.Command, args []string) (err error) {
	cmd.SilenceUsage = true

	dataDir := viper.GetString("global-data-dir")
	userLog.Debug("fireacme binary started", zap.String("data_dir", dataDir))

	configFile := viper.GetString("global-config-file")
	userLog.Printf("Starting StreamingFast on Acme with config file '%s'", configFile)

	err = Start(configFile, dataDir, args)
	if err != nil {
		return fmt.Errorf("unable to launch: %w", err)
	}

	// If an error occurred, saying Goodbye is not greate
	userLog.Printf("Goodbye")
	return
}

func Start(configFile string, dataDir string, args []string) (err error) {
	dataDirAbs, err := filepath.Abs(dataDir)
	if err != nil {
		return fmt.Errorf("unable to setup directory structure: %w", err)
	}

	err = makeDirs([]string{dataDirAbs})
	if err != nil {
		return err
	}

	//meshClient, err := dmeshClient.New(viper.GetString("search-common-mesh-dsn"))
	//if err != nil {
	//	return fmt.Errorf("unable to create dmesh client: %w", err)
	//}

	tracker := bstream.NewTracker(50)

	//blockmetaAddr := viper.GetString("common-blockmeta-addr")
	//if blockmetaAddr != "" {
	//	conn, err := dgrpc.NewInternalClient(blockmetaAddr)
	//	if err != nil {
	//		userLog.Warn("cannot get grpc connection to blockmeta, disabling this startBlockResolver for search indexer", zap.Error(err), zap.String("blockmeta_addr", blockmetaAddr))
	//	} else {
	//		blockmetaCli := pbblockmeta.NewBlockIDClient(conn)
	//		tracker.AddResolver(pbblockmeta.StartBlockResolver(blockmetaCli))
	//	}
	//}

	tracker.AddResolver(bstream.OffsetStartBlockResolver(200))

	modules := &launcher.Runtime{
		//SearchDmeshClient: meshClient,
		AbsDataDir: dataDirAbs,
		Tracker:    tracker,
	}

	atmCacheEnabled := viper.GetBool("common-atm-cache-enabled")
	if atmCacheEnabled {
		bstream.GetBlockPayloadSetter = bstream.ATMCachedPayloadSetter

		cacheDir := MustReplaceDataDir(modules.AbsDataDir, viper.GetString("common-atm-cache-dir"))
		storeUrl := MustReplaceDataDir(modules.AbsDataDir, viper.GetString("common-blocks-store-url"))
		maxRecentEntryBytes := viper.GetInt("common-atm-max-recent-entry-bytes")
		maxEntryByAgeBytes := viper.GetInt("common-atm-max-entry-by-age-bytes")
		bstream.InitCache(storeUrl, cacheDir, maxRecentEntryBytes, maxEntryByAgeBytes)
	}

	bstream.GetProtocolFirstStreamableBlock = uint64(viper.GetInt("common-first-streamable-block"))

	err = bstream.ValidateRegistry()
	if err != nil {
		return fmt.Errorf("protocol specific hooks not configured correctly: %w", err)
	}

	launch := launcher.NewLauncher(modules)
	userLog.Debug("launcher created")

	runByDefault := func(app string) bool {
		if app == "search-forkresolver" {
			return false
		}
		return true
	}

	apps := launcher.ParseAppsFromArgs(args, runByDefault)
	if len(args) == 0 {
		apps = launcher.ParseAppsFromArgs(launcher.DfuseConfig["start"].Args, runByDefault)
	}
	userLog.Printf("Launching applications: %s", strings.Join(apps, ","))
	if err = launch.Launch(apps); err != nil {
		return err
	}

	signalHandler := derr.SetupSignalHandler(viper.GetDuration("common-system-shutdown-signal-delay"))
	select {
	case <-signalHandler:
		userLog.Printf("Received termination signal, quitting")
		go launch.Close()
	case appID := <-launch.Terminating():
		if launch.Err() == nil {
			userLog.Printf("Application %s triggered a clean shutdown, quitting", appID)
		} else {
			userLog.Printf("Application %s shutdown unexpectedly, quitting", appID)
			err = launch.Err()
		}
	}

	launch.WaitForTermination()

	return
}
