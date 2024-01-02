package main

import (
	"errors"
	"fmt"
	pbantelope "github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/spf13/cobra"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/dstore"
	firecore "github.com/streamingfast/firehose-core"
	"go.uber.org/zap"
	"io"
)

func newCheckBlocksCmd(logger *zap.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "check-blocks <src-blocks-store> <start-block> <stop-block>",
		Short: "checks all blocks for decoding issues (detects the invalid UTF-8 issue).",
		Args:  cobra.ExactArgs(3),
		RunE:  createCheckBlocksE(logger),
	}
}

func createCheckBlocksE(logger *zap.Logger) firecore.CommandExecutor {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		srcStore, err := dstore.NewDBinStore(args[0])
		if err != nil {
			return fmt.Errorf("unable to create source store: %w", err)
		}

		start := mustParseUint64(args[1])
		stop := mustParseUint64(args[2])

		if stop <= start {
			return fmt.Errorf("stop block must be greater than start block")
		}

		logger.Info("checking antelope blocks", zap.Uint64("start_block", start), zap.Uint64("stop_block", stop))

		lastFileProcessed := ""
		startWalkFrom := fmt.Sprintf("%010d", start-(start%100))
		err = srcStore.WalkFrom(ctx, "", startWalkFrom, func(filename string) error {
			logger.Debug("checking merged block file", zap.String("filename", filename))

			startBlock := mustParseUint64(filename)

			if startBlock > stop {
				logger.Debug("stopping at merged block file above stop block", zap.String("filename", filename), zap.Uint64("stop", stop))
				return io.EOF
			}

			if startBlock+100 < start {
				logger.Debug("skipping merged block file below start block", zap.String("filename", filename))
				return nil
			}

			rc, err := srcStore.OpenObject(ctx, filename)
			if err != nil {
				return fmt.Errorf("failed to open %s: %w", filename, err)
			}
			defer rc.Close()

			br, err := bstream.NewDBinBlockReader(rc)
			if err != nil {
				return fmt.Errorf("creating block reader: %w", err)
			}

			i := 0
			for {
				block, err := br.Read()
				if err == io.EOF {
					break
				}

				var antelopeBlock pbantelope.Block
				err = block.Payload.UnmarshalTo(&antelopeBlock)
				if err != nil {
					fmt.Printf("block_num: %d - unable to decode: %s\n", block.Number, err)
				} else {
					logger.Debug("successfully decoded antelope block", zap.Any("block_num", block.Number))
				}

				i++
			}

			if i != 100 {
				return fmt.Errorf("expected to have read 100 blocks, we have read %d. Bailing out.", i)
			}

			lastFileProcessed = filename
			logger.Info("finished merged block", zap.String("filename", filename))

			return nil
		})
		fmt.Printf("Last file processed: %s.dbin.zst\n", lastFileProcessed)

		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return err
		}

		return nil
	}
}
