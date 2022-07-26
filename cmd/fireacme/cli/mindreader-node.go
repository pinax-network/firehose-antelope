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
	"github.com/streamingfast/bstream/blockstream"
	"github.com/streamingfast/firehose-acme/nodemanager/codec"
	"github.com/streamingfast/logging"
	nodeManager "github.com/streamingfast/node-manager"
	"github.com/streamingfast/node-manager/mindreader"
	"go.uber.org/zap"
)

func init() {
	registerNode("extractor", registerExtractorNodeFlags, ExtractorNodeManagerAPIAddr)
}

func registerExtractorNodeFlags(cmd *cobra.Command) error {
	cmd.Flags().String("extractor-node-grpc-listen-addr", ExtractorNodeGRPCAddr, "The gRPC listening address to use for serving real-time blocks")
	cmd.Flags().Bool("extractor-node-discard-after-stop-num", false, "Ignore remaining blocks being processed after stop num (only useful if we discard the extractor data after reprocessing a chunk of blocks)")
	cmd.Flags().String("extractor-node-working-dir", "{data-dir}/extractor/work", "Path where extractor will stores its files")
	cmd.Flags().Uint("extractor-node-start-block-num", 0, "Blocks that were produced with smaller block number then the given block num are skipped")
	cmd.Flags().Uint("extractor-node-stop-block-num", 0, "Shutdown extractor when we the following 'stop-block-num' has been reached, inclusively.")
	cmd.Flags().Int("extractor-node-blocks-chan-capacity", 100, "Capacity of the channel holding blocks read by the extractor. Process will shutdown superviser/geth if the channel gets over 90% of that capacity to prevent horrible consequences. Raise this number when processing tiny blocks very quickly")
	cmd.Flags().String("extractor-node-one-block-suffix", "default", FlagDescription(`
		Unique identifier for extractor, so that it can produce 'oneblock files' in the same store as another instance without competing
		for writes. You should set this flag if you have multiple extractor running, each one should get a unique identifier, the
		hostname value is a good value to use.
	`))

	return nil
}

func getMindreaderLogPlugin(
	blockStreamServer *blockstream.Server,
	mergedBlockStoreURL string,
	workingDir string,
	batchStartBlockNum uint64,
	batchStopBlockNum uint64,
	blocksChanCapacity int,
	oneBlockFileSuffix string,
	operatorShutdownFunc func(error),
	metricsAndReadinessManager *nodeManager.MetricsAndReadinessManager,
	appLogger *zap.Logger,
	appTracer logging.Tracer,
) (*mindreader.MindReaderPlugin, error) {
	consoleReaderFactory := func(lines chan string) (mindreader.ConsolerReader, error) {
		return codec.NewConsoleReader(appLogger, lines)
	}

	return mindreader.NewMindReaderPlugin(
		mergedBlockStoreURL,
		workingDir,
		consoleReaderFactory,
		batchStartBlockNum,
		batchStopBlockNum,
		blocksChanCapacity,
		metricsAndReadinessManager.UpdateHeadBlock,
		func(error) {
			operatorShutdownFunc(nil)
		},
		oneBlockFileSuffix,
		blockStreamServer,
		appLogger,
		appTracer,
	)
}
