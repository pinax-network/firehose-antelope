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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	pbantelope "github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/dstore"
	"go.uber.org/zap"
)

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Prints of one block or merged blocks file",
}

var oneBlockCmd = &cobra.Command{
	Use:   "one-block {block_num}",
	Short: "Prints a block from a one-block file",
	Args:  cobra.ExactArgs(1),
	RunE:  printOneBlockE,
}

var blocksCmd = &cobra.Command{
	Use:   "blocks {block_num}",
	Short: "Prints the content summary of merged blocks file",
	Args:  cobra.ExactArgs(1),
	RunE:  printBlocksE,
}

var blockCmd = &cobra.Command{
	Use:   "block {block_num}",
	Short: "Finds and prints one block from merged blocks file",
	Args:  cobra.ExactArgs(1),
	RunE:  printBlockE,
}

func init() {
	Cmd.AddCommand(printCmd)

	printCmd.AddCommand(oneBlockCmd)

	printCmd.AddCommand(blocksCmd)
	blocksCmd.PersistentFlags().Bool("transactions", false, "Include transaction IDs in output")

	printCmd.AddCommand(blockCmd)
	blockCmd.Flags().String("transaction", "", "Filters transaction by this hash")

	printCmd.PersistentFlags().Uint64("transactions-for-block", 0, "Include transaction IDs in output")
	printCmd.PersistentFlags().Bool("transactions", false, "Include transaction IDs in output")
	printCmd.PersistentFlags().Bool("calls", false, "Include transaction's Call data in output")
	printCmd.PersistentFlags().Bool("instructions", false, "Include instruction output")
	printCmd.PersistentFlags().String("store", "", "block store")
}

func printBlocksE(cmd *cobra.Command, args []string) error {
	printTransactions := viper.GetBool("transactions")

	blockNum, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse block number %q: %w", args[0], err)
	}

	str := viper.GetString("store")

	store, err := dstore.NewDBinStore(str)
	if err != nil {
		return fmt.Errorf("unable to create store at path %q: %w", store, err)
	}

	filename := fmt.Sprintf("%010d", blockNum)
	reader, err := store.OpenObject(context.Background(), filename)
	if err != nil {
		fmt.Printf("❌ Unable to read blocks filename %s: %s\n", filename, err)
		return err
	}
	defer reader.Close()

	readerFactory, err := bstream.GetBlockReaderFactory.New(reader)
	if err != nil {
		fmt.Printf("❌ Unable to read blocks filename %s: %s\n", filename, err)
		return err
	}

	seenBlockCount := 0
	for {
		block, err := readerFactory.Read()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Total blocks: %d\n", seenBlockCount)
				return nil
			}
			return fmt.Errorf("error receiving blocks: %w", err)
		}

		seenBlockCount++

		antelopeBlock := block.ToProtocol().(*pbantelope.Block)

		fmt.Printf("Block #%d (%s) (prev: %s): %d transactions\n",
			block.Num(),
			block.ID()[0:7],
			block.PreviousID()[0:7],
			len(antelopeBlock.Transactions()),
		)
		if printTransactions {
			fmt.Println("- Transactions: ")
			for _, t := range antelopeBlock.Transactions() {
				fmt.Println("  * ", t.Id)
			}
			fmt.Println()
		}
	}
}

func printBlockE(cmd *cobra.Command, args []string) error {
	printTransactions := viper.GetBool("transactions")
	transactionFilter := viper.GetString("transaction")

	zlog.Info("printing block",
		zap.Bool("print_transactions", printTransactions),
		zap.String("transaction_filter", transactionFilter),
	)

	blockNum, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse block number %q: %w", args[0], err)
	}

	str := viper.GetString("store")

	store, err := dstore.NewDBinStore(str)
	if err != nil {
		return fmt.Errorf("unable to create store at path %q: %w", store, err)
	}

	mergedBlockNum := blockNum - (blockNum % 100)
	zlog.Info("finding merged block file",
		zap.Uint64("merged_block_num", mergedBlockNum),
		zap.Uint64("block_num", blockNum),
	)

	filename := fmt.Sprintf("%010d", mergedBlockNum)
	reader, err := store.OpenObject(context.Background(), filename)
	if err != nil {
		fmt.Printf("❌ Unable to read blocks filename %s: %s\n", filename, err)
		return err
	}
	defer reader.Close()

	readerFactory, err := bstream.GetBlockReaderFactory.New(reader)
	if err != nil {
		fmt.Printf("❌ Unable to read blocks filename %s: %s\n", filename, err)
		return err
	}

	for {
		block, err := readerFactory.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("error reading blocks: %w", err)
		}

		if block.Number != blockNum {
			zlog.Debug("skipping block",
				zap.Uint64("desired_block_num", blockNum),
				zap.Uint64("block_num", block.Number),
			)
			continue
		}
		antelopeBlock := block.ToProtocol().(*pbantelope.Block)

		fmt.Printf("Block #%d (%s) (prev: %s): %d transactions\n",
			block.Num(),
			block.ID()[0:7],
			block.PreviousID()[0:7],
			len(antelopeBlock.Transactions()),
		)
		if printTransactions {
			fmt.Println("- Transactions: ")
			for _, t := range antelopeBlock.Transactions() {
				fmt.Printf("  * %s\n", t.Id)
			}
		}
		continue
	}
}

func printOneBlockE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	blockNum, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse block number %q: %w", args[0], err)
	}

	str := viper.GetString("store")

	store, err := dstore.NewDBinStore(str)
	if err != nil {
		return fmt.Errorf("unable to create store at path %q: %w", store, err)
	}

	var files []string
	filePrefix := fmt.Sprintf("%010d", blockNum)
	err = store.Walk(ctx, filePrefix, func(filename string) (err error) {
		files = append(files, filename)
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to find on block files: %w", err)
	}

	for _, filepath := range files {
		reader, err := store.OpenObject(ctx, filepath)
		if err != nil {
			fmt.Printf("❌ Unable to read block filename %s: %s\n", filepath, err)
			return err
		}
		defer reader.Close()

		readerFactory, err := bstream.GetBlockReaderFactory.New(reader)
		if err != nil {
			fmt.Printf("❌ Unable to read blocks filename %s: %s\n", filepath, err)
			return err
		}

		//fmt.Printf("One Block File: %s\n", store.ObjectURL(filepath))

		block, err := readerFactory.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("reading block: %w", err)
		}

		if err = printBlock(block); err != nil {
			return err
		}

	}
	return nil
}

func printBlock(block *bstream.Block) error {
	nativeBlock := block.ToProtocol().(*pbantelope.Block)

	data, err := json.MarshalIndent(nativeBlock, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshall: %w", err)
	}

	fmt.Println(string(data))

	return nil
}
