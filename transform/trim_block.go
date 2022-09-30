package transform

import (
	"github.com/EOS-Nation/firehose-antelope/types"
	pbantelope "github.com/EOS-Nation/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/bstream/transform"
)

var TrimBlock = &transform.Factory{
	Obj:     nil,
	NewFunc: nil,
}

type BlockTrimmer struct{}

func (b *BlockTrimmer) Transform(readOnlyBlk *bstream.Block, in transform.Input) (transform.Output, error) {

	decodedBlock, err := types.BlockDecoder(readOnlyBlk)
	if err != nil {
		return nil, err
	}

	// We need to create a new instance because this block could be in the live segment
	// which is shared across all streams that requires live block. As such, we cannot modify
	// them in-place, so we require to create a new instance.
	//
	// The copy is mostly shallow since we copy over pointers element but some part are deep
	// copied like ActionTrace which requires trimming.
	fullBlock := decodedBlock.(*pbantelope.Block)
	block := &pbantelope.Block{
		Id:                       fullBlock.Id,
		Number:                   fullBlock.Number,
		DposIrreversibleBlocknum: fullBlock.DposIrreversibleBlocknum,
		Header: &pbantelope.BlockHeader{
			Timestamp: fullBlock.Header.Timestamp,
			Producer:  fullBlock.Header.Producer,
		},
	}

	var newTrace func(fullTrxTrace *pbantelope.TransactionTrace) (trxTrace *pbantelope.TransactionTrace)
	newTrace = func(fullTrxTrace *pbantelope.TransactionTrace) (trxTrace *pbantelope.TransactionTrace) {
		trxTrace = &pbantelope.TransactionTrace{
			Id:        fullTrxTrace.Id,
			Receipt:   fullTrxTrace.Receipt,
			Scheduled: fullTrxTrace.Scheduled,
			Exception: fullTrxTrace.Exception,
		}

		if fullTrxTrace.FailedDtrxTrace != nil {
			trxTrace.FailedDtrxTrace = newTrace(fullTrxTrace.FailedDtrxTrace)
		}

		trxTrace.ActionTraces = make([]*pbantelope.ActionTrace, len(fullTrxTrace.ActionTraces))
		for i, fullActTrace := range fullTrxTrace.ActionTraces {
			actTrace := &pbantelope.ActionTrace{
				Receiver:                               fullActTrace.Receiver,
				ContextFree:                            fullActTrace.ContextFree,
				Exception:                              fullActTrace.Exception,
				ErrorCode:                              fullActTrace.ErrorCode,
				ActionOrdinal:                          fullActTrace.ActionOrdinal,
				CreatorActionOrdinal:                   fullActTrace.CreatorActionOrdinal,
				ClosestUnnotifiedAncestorActionOrdinal: fullActTrace.ClosestUnnotifiedAncestorActionOrdinal,
				ExecutionIndex:                         fullActTrace.ExecutionIndex,
			}

			if fullActTrace.Action != nil {
				actTrace.Action = &pbantelope.Action{
					Account:       fullActTrace.Action.Account,
					Name:          fullActTrace.Action.Name,
					Authorization: fullActTrace.Action.Authorization,
					JsonData:      fullActTrace.Action.JsonData,
				}

				if fullActTrace.Action.JsonData == "" {
					actTrace.Action.RawData = fullActTrace.Action.RawData
				}
			}

			if fullActTrace.Receipt != nil {
				actTrace.Receipt = &pbantelope.ActionReceipt{
					GlobalSequence: fullActTrace.Receipt.GlobalSequence,
				}
			}

			trxTrace.ActionTraces[i] = actTrace
		}

		return trxTrace
	}

	traces := make([]*pbantelope.TransactionTrace, len(fullBlock.TransactionTraces()))
	for i, fullTrxTrace := range fullBlock.TransactionTraces() {
		traces[i] = newTrace(fullTrxTrace)
	}

	if fullBlock.FilteringApplied {
		block.FilteredTransactionTraces = traces
	} else {
		block.UnfilteredTransactionTraces = traces
	}

	return block, nil
}
