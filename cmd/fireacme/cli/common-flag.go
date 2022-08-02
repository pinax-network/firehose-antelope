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
	"go.uber.org/zap"
)

func init() {
	launcher.RegisterCommonFlags = func(_ *zap.Logger, cmd *cobra.Command) error {
		//Common stores configuration flags
		cmd.Flags().String("common-one-blocks-store-url", OneBlockStoreURL, "[COMMON] Store URL to read/write one-block files, use by extractor, merger")
		cmd.Flags().String("common-merged-blocks-store-url", MergedBlocksStoreURL, "[COMMON] Store URL where to read/write merged blocks, used by: extractor, merger, firehose.")
		cmd.Flags().String("common-relayer-addr", RelayerServingAddr, "[COMMON] gRPC endpoint to get real-time blocks, used by: firehose")

		cmd.Flags().Bool("common-blocks-cache-enabled", false, FlagDescription(`
			[COMMON] Use a disk cache to store the blocks data to disk and instead of keeping it in RAM. By enabling this, block's Protobuf content, in bytes,
			is kept on file system instead of RAM. This is done as soon the block is downloaded from storage. This is a tradeoff between RAM and Disk, if you
			are going to serve only a handful of concurrent requests, it's suggested to keep is disabled, if you encounter heavy RAM consumption issue, specially
			by the firehose component, it's definitely a good idea to enable it and configure it properly through the other 'common-blocks-cache-...' flags. The cache is
			split in two portions, one keeping N total bytes of blocks of the most recently used blocks and the other one keeping the N earliest blocks as
			requested by the various consumers of the cache.
		`))
		cmd.Flags().String("common-blocks-cache-dir", BlocksCacheDirectory, FlagDescription(`
			[COMMON] Blocks cache directory where all the block's bytes will be cached to disk instead of being kept in RAM.
			This should be a disk that persists across restarts of the Firehose component to reduce the the strain on the disk
			when restarting and streams reconnects. The size of disk must at least big (with a 10% buffer) in bytes as the sum of flags'
			value for  'common-blocks-cache-max-recent-entry-bytes' and 'common-blocks-cache-max-entry-by-age-bytes'.
		`))
		cmd.Flags().Int("common-blocks-cache-max-recent-entry-bytes", 21474836480, FlagDescription(`
			[COMMON] Blocks cache max size in bytes of the most recently used blocks, after the limit is reached, blocks are evicted from the cache.
		`))
		cmd.Flags().Int("common-blocks-cache-max-entry-by-age-bytes", 21474836480, FlagDescription(`
			[COMMON] Blocks cache max size in bytes of the earliest used blocks, after the limit is reached, blocks are evicted from the cache.
		`))

		cmd.Flags().Int("common-first-streamable-block", FirstStreamableBlock, "[COMMON] First streamable block of the chain, ")

		// Authentication, metering and rate limiter plugins
		cmd.Flags().String("common-auth-plugin", "null://", "[COMMON] Auth plugin URI, see streamingfast/dauth repository")
		cmd.Flags().String("common-metering-plugin", "null://", "[COMMON] Metering plugin URI, see streamingfast/dmetering repository")

		// System Behavior
		cmd.Flags().Duration("common-system-shutdown-signal-delay", 0, FlagDescription(`
			[COMMON] Add a delay between receiving SIGTERM signal and shutting down apps.
			Apps will respond negatively to /healthz during this period
		`))
		return nil
	}
}
