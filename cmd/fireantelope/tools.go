package main

import (
	"fmt"
	pbantelope "github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/streamingfast/bstream"
	"io"
)

func printBlock(block *bstream.Block, alsoPrintTransactions bool, out io.Writer) error {
	nativeBlock := block.ToProtocol().(*pbantelope.Block)

	if _, err := fmt.Fprintf(out, "Block #%d (%s) (prev: %s): %d transactions\n",
		block.Num(),
		block.ID(),
		block.PreviousID()[0:7],
		len(nativeBlock.Transactions()),
	); err != nil {
		return err
	}

	if alsoPrintTransactions {
		for _, trx := range nativeBlock.Transactions() {
			if _, err := fmt.Fprintf(out, "- Transaction %d\n", trx.Id); err != nil {
				return err
			}
		}
	}

	//data, err := json.MarshalIndent(nativeBlock, "", "  ")
	//if err != nil {
	//	return fmt.Errorf("json marshall: %w", err)
	//}
	//
	//fmt.Println(string(data))

	return nil
}
