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

package tools

import (
	"fmt"

	pbantelope "github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	sftools "github.com/streamingfast/sf-tools"
)

// CmdCheck is used in sf-ethereum-priv where additional checks are added.
var CheckCmd = &cobra.Command{Use: "check", Short: "Various checks for deployment, data integrity & debugging"}

var checkMergedBlocksCmd = &cobra.Command{
	// TODO: Not sure, it's now a required thing, but we could probably use the same logic as `start`
	//       and avoid altogether passing the args. If this would also load the config and everything else,
	//       that would be much more seamless!
	Use:   "merged-blocks {store-url}",
	Short: "Checks for any holes in merged blocks as well as ensuring merged blocks integrity",
	Args:  cobra.ExactArgs(1),
	RunE:  checkMergedBlocksE,
	Example: ExamplePrefixed("sfeth tools check merged-blocks", `
		"./sf-data/storage/merged-blocks"
		"gs://<project>/<bucket>/<path> -s"
		"s3://<project>/<bucket>/<path> -f"
		"az://<project>/<bucket>/<path> -r \"10 000 - 1 000 000"
	`),
}

func init() {
	Cmd.AddCommand(CheckCmd)
	CheckCmd.AddCommand(checkMergedBlocksCmd)

	CheckCmd.PersistentFlags().StringP("range", "r", "", "Block range to use for the check")
	CheckCmd.PersistentFlags().IntP("workers", "w", 1, "Number of workers")
	CheckCmd.PersistentFlags().IntP("batch", "b", 25000, "batch size")

	checkMergedBlocksCmd.Flags().BoolP("print-stats", "s", false, "Natively decode each block in the segment and print statistics about it, ensuring it contains the required blocks")
	checkMergedBlocksCmd.Flags().BoolP("print-full", "f", false, "Natively decode each block and print the full JSON representation of the block, should be used with a small range only if you don't want to be overwhelmed")
}

func checkMergedBlocksE(cmd *cobra.Command, args []string) error {
	storeURL := args[0]
	fileBlockSize := uint32(100)

	blockRange, err := sftools.Flags.GetBlockRange("range")
	if err != nil {
		return err
	}

	printDetails := sftools.PrintNothing
	if viper.GetBool("print-stats") {
		printDetails = sftools.PrintStats
	}

	if viper.GetBool("print-full") {
		printDetails = sftools.PrintFull
	}

	batchSize := viper.GetInt("batch")
	workers := viper.GetInt("workers")

	if workers == 1 {
		return sftools.CheckMergedBlocks(cmd.Context(), zlog, storeURL, fileBlockSize, blockRange, blockPrinter, printDetails)
	} else {
		return sftools.CheckMergedBlocksBatch(cmd.Context(), zlog, storeURL, fileBlockSize, blockRange, blockPrinter, printDetails, batchSize, workers)
	}
}

func blockPrinter(block *bstream.Block) {
	dummBlock := block.ToProtocol().(*pbantelope.Block)
	fmt.Printf("Block %s, Prev: %s: %v timestamp\n",
		block.AsRef(),
		block.PreviousID(),
		dummBlock.Time(),
	)
}
