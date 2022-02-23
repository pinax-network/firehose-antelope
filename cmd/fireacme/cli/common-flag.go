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
	"github.com/spf13/cobra"
	"github.com/streamingfast/dlauncher/launcher"
)

func init() {
	launcher.RegisterCommonFlags = func(cmd *cobra.Command) error {
		//Common stores configuration flags
		cmd.Flags().String("common-blocks-store-url", MergedBlocksStoreURL, "[COMMON] Store URL (with prefix) where to read/write. Used by: relayer, fluxdb, trxdb-loader, blockmeta, search-indexer, search-live, search-forkresolver")
		cmd.Flags().String("common-oneblock-store-url", OneBlockStoreURL, "[COMMON] Store URL (with prefix) to read/write one-block files. Used by: mindreader, merger")
		cmd.Flags().String("common-blockstream-addr", RelayerServingAddr, "[COMMON] gRPC endpoint to get real-time blocks. Used by: fluxdb, trxdb-loader, blockmeta, search-indexer, search-live (relayer uses its own --relayer-blockstream-addr)")

		cmd.Flags().Bool("common-atm-cache-enabled", false, "[COMMON] enable ATM caching")
		cmd.Flags().String("common-atm-cache-dir", ATMDirectory, "[COMMON] ATM cache file directory.")
		cmd.Flags().Int("common-atm-max-recent-entry-bytes", 21474836480, "[COMMON] ATM cache max size in bytes of recent entry heap")
		cmd.Flags().Int("common-atm-max-entry-by-age-bytes", 21474836480, "[COMMON] ATM cache max size in bytes of age entry heap")

		cmd.Flags().Int("common-first-streamable-block", FirstStreamableBlock, "[COMMON] first streamable block number")

		// Authentication, metering and rate limiter plugins
		cmd.Flags().String("common-auth-plugin", "null://", "[COMMON] Auth plugin URI, see dfuse-io/dauth repository")
		cmd.Flags().String("common-metering-plugin", "null://", "[COMMON] Metering plugin URI, see dfuse-io/dmetering repository")

		// System Behavior
		cmd.Flags().Duration("common-system-shutdown-signal-delay", 0, "[COMMON] Add a delay between receiving SIGTERM signal and shutting down apps. Apps will respond negatively to /healthz during this period")
		return nil
	}
}
