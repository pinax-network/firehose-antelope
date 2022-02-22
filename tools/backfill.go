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
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/streamingfast/dbin"
	"github.com/streamingfast/dstore"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	pbcodec "github.com/streamingfast/sf-near/pb/sf/near/codec/v1"
	sftools "github.com/streamingfast/sf-tools"
	"go.uber.org/zap"
)

var numberRegex = regexp.MustCompile(`(\d{10})`)

var BackfillCmd = &cobra.Command{Use: "backfill", Short: "Various tools for updating merged block files"}

var backfillPrevHeightCmd = &cobra.Command{
	Use:   "prev-height {input-store-url} {output-store-url}",
	Short: "update merge block files to set previous height data",
	Args:  cobra.ExactArgs(2),
	RunE:  backfillPrevHeightE,
}

var backfillPrevHeightCheckCmd = &cobra.Command{
	Use:   "prev-height-check {input-store-url}",
	Short: "check merge block files to validate previous height data",
	Args:  cobra.ExactArgs(1),
	RunE:  backfillPrevHeightCheckE,
}

func init() {
	Cmd.AddCommand(BackfillCmd)
	BackfillCmd.AddCommand(backfillPrevHeightCmd)
	BackfillCmd.AddCommand(backfillPrevHeightCheckCmd)

	BackfillCmd.PersistentFlags().StringP("range", "r", "", "Block range to use for the check")
}

var errStopWalk = fmt.Errorf("stop walk")

func backfillPrevHeightE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	blockRange, err := sftools.Flags.GetBlockRange("range")
	if err != nil {
		return fmt.Errorf("error parsing block range: %w", err)
	}

	fileBlockSize := 100
	walkPrefix := sftools.WalkBlockPrefix(blockRange, uint32(fileBlockSize))

	inputBlocksStore, err := dstore.NewDBinStore(args[0])
	if err != nil {
		return fmt.Errorf("error opening input store %s: %w", args[0], err)
	}

	var baseNum32 uint32
	heightMap := make(map[string]uint64)

	err = inputBlocksStore.Walk(ctx, walkPrefix, ".tmp", func(filename string) (err error) {
		match := numberRegex.FindStringSubmatch(filename)
		if match == nil {
			zlog.Debug("file does not match pattern", zap.String("filename", filename))
			return nil
		}

		baseNum, _ := strconv.ParseUint(match[1], 10, 32)
		if baseNum+uint64(fileBlockSize)-1 < blockRange.Start {
			zlog.Debug("file is before block range start", zap.String("filename", filename), zap.Uint64("block_range_start", blockRange.Start))
			return nil
		}
		baseNum32 = uint32(baseNum)

		try := 0
		var obj io.ReadCloser
		for {
			try += 1
			obj, err = inputBlocksStore.OpenObject(ctx, filename)
			if err != nil {
				if try > 10 {
					return fmt.Errorf("error reading file %s from input store %s: %w", filename, args[0], err)
				}
				time.Sleep(time.Duration(try) * time.Second)
				continue
			}
			break
		}

		binReader := dbin.NewReader(obj)
		contentType, version, err := binReader.ReadHeader()
		if err != nil {
			return fmt.Errorf("error reading file header for file %s: %w", filename, err)
		}
		defer binReader.Close()

		buffer := bytes.NewBuffer(nil)
		binWriter := dbin.NewWriter(buffer)
		defer binWriter.Close()

		err = binWriter.WriteHeader(contentType, int(version))
		if err != nil {
			return fmt.Errorf("error writing block header: %w", err)
		}

		firstSeenBlock := true
		for {
			line, err := binReader.ReadMessage()
			if err == io.EOF {
				zlog.Debug("eof", zap.String("filename", filename))
				break
			}

			if err != nil {
				return fmt.Errorf("reading block: %w", err)
			}

			if len(line) == 0 {
				zlog.Debug("empty line", zap.String("filename", filename))
				break
			}

			// decode block data
			bstreamBlock := new(pbbstream.Block)
			err = proto.Unmarshal(line, bstreamBlock)
			if err != nil {
				return fmt.Errorf("unmarshaling block proto: %w", err)
			}

			blockBytes := bstreamBlock.GetPayloadBuffer()

			block := new(pbcodec.Block)
			err = proto.Unmarshal(blockBytes, block)
			if err != nil {
				return fmt.Errorf("unmarshaling block proto: %w", err)
			}

			// save current id/height
			heightMap[block.ID()] = block.Number()

			prevHeight, ok := heightMap[block.PreviousID()]
			if !ok {
				if firstSeenBlock {
					firstSeenBlock = false
					zlog.Debug("skipping first block update. no prev_height data yet", zap.Uint64("block", block.Number()))
				} else {
					return fmt.Errorf("could not find previous height for block id %s", block.ID())
				}
			} else {
				// update current block prev_height
				block.Header.PrevHeight = prevHeight
				zlog.Debug("updated prev_height", zap.Uint64("block", block.Number()), zap.Uint64("prev_height", prevHeight))
			}

			// encode block data
			backFilledBlockBytes, err := proto.Marshal(block)
			if err != nil {
				return fmt.Errorf("marshaling block proto: %w", err)
			}

			bstreamBlock.PayloadBuffer = backFilledBlockBytes

			bstreamBlockBytes, err := proto.Marshal(bstreamBlock)
			if err != nil {
				return fmt.Errorf("marshaling bstream block: %w", err)
			}

			err = binWriter.WriteMessage(bstreamBlockBytes)
			if err != nil {
				return fmt.Errorf("error writing block: %w", err)
			}
		}

		err = obj.Close()
		if err != nil {
			return fmt.Errorf("error closing object %s: %w", filename, err)
		}

		outputBlocksStore, err := dstore.NewDBinStore(args[1])
		if err != nil {
			return fmt.Errorf("error opening output store %s: %w", args[0], err)
		}

		outputBlocksStore.SetOverwrite(true)

		try = 0
		for {
			try += 1
			err = outputBlocksStore.WriteObject(ctx, filename, buffer)
			if err != nil {
				if try > 10 {
					return fmt.Errorf("error writing file %s to output store %s: %w", filename, args[1], err)
				}
				time.Sleep(time.Duration(try) * time.Second)
				continue
			}
			zlog.Debug("saved output file", zap.String("store", outputBlocksStore.BaseURL().String()), zap.String("filename", filename))
			break
		}

		// check range upper bound
		if !blockRange.Unbounded() {
			roundedEndBlock := sftools.RoundToBundleEndBlock(baseNum32, uint32(fileBlockSize))
			if roundedEndBlock >= uint32(blockRange.Stop-1) {
				return errStopWalk
			}
		}

		zlog.Info("updated blocks", zap.String("file", filename))
		return nil
	})

	if err != nil && err != errStopWalk {
		return err
	}

	return nil
}

func backfillPrevHeightCheckE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	blockRange, err := sftools.Flags.GetBlockRange("range")
	if err != nil {
		return fmt.Errorf("error parsing block range: %w", err)
	}

	fileBlockSize := 100
	walkPrefix := sftools.WalkBlockPrefix(blockRange, uint32(fileBlockSize))

	inputBlocksStore, err := dstore.NewDBinStore(args[0])
	if err != nil {
		return fmt.Errorf("error opening input store %s: %w", args[0], err)
	}

	var baseNum32 uint32
	firstBlockSeen := true

	err = inputBlocksStore.Walk(ctx, walkPrefix, ".tmp", func(filename string) (err error) {
		match := numberRegex.FindStringSubmatch(filename)
		if match == nil {
			return nil
		}

		baseNum, _ := strconv.ParseUint(match[1], 10, 32)
		if baseNum+uint64(fileBlockSize)-1 < blockRange.Start {
			return nil
		}
		baseNum32 = uint32(baseNum)

		zlog.Debug("checking file", zap.String("filename", filename))

		obj, err := inputBlocksStore.OpenObject(ctx, filename)
		if err != nil {
			return fmt.Errorf("error reading file %s from input store %s: %w", filename, args[0], err)
		}

		binReader := dbin.NewReader(obj)
		_, _, err = binReader.ReadHeader()
		if err != nil {
			if err == io.EOF {
				zlog.Info("eof reading file header", zap.String("filename", filename))
				return nil
			}
			return fmt.Errorf("error reading file header for file %s: %w", filename, err)
		}
		defer binReader.Close()

		for {
			line, err := binReader.ReadMessage()
			if err == io.EOF {
				break
			}

			if err != nil {
				return fmt.Errorf("reading block: %w", err)
			}

			if len(line) == 0 {
				break
			}

			// decode block data
			bstreamBlock := new(pbbstream.Block)
			err = proto.Unmarshal(line, bstreamBlock)
			if err != nil {
				return fmt.Errorf("unmarshaling block proto: %w", err)
			}

			blockBytes := bstreamBlock.GetPayloadBuffer()

			block := new(pbcodec.Block)
			err = proto.Unmarshal(blockBytes, block)
			if err != nil {
				return fmt.Errorf("unmarshaling block proto: %w", err)
			}

			prevHeight := block.Header.PrevHeight
			if prevHeight == 0 {
				if firstBlockSeen {
					zlog.Debug("first block, skipping check", zap.Uint64("block", block.Number()))
					firstBlockSeen = false
					continue
				}
				return fmt.Errorf("previous height not set for block number %d in file %s", block.Number(), filename)
			}
		}

		err = obj.Close()
		if err != nil {
			return fmt.Errorf("error closing object %s: %w", filename, err)
		}

		// check range upper bound
		if !blockRange.Unbounded() {
			roundedEndBlock := sftools.RoundToBundleEndBlock(baseNum32, uint32(fileBlockSize))
			if roundedEndBlock >= uint32(blockRange.Stop-1) {
				return errStopWalk
			}
		}

		zlog.Info("checked file", zap.String("file", filename))
		return nil
	})

	zlog.Debug("file walk ended", zap.Error(err))

	if err != nil && err != errStopWalk {
		return err
	}

	return nil
}
