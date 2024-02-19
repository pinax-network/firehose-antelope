package main

import (
	"fmt"
	pbantelope "github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	pbbstream "github.com/streamingfast/bstream/pb/sf/bstream/v1"
	firecore "github.com/streamingfast/firehose-core"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"time"
)

var fixedTimestamp = timestamppb.New(time.Date(2006, 01, 02, 15, 04, 05, 0, time.UTC))

func mustParseUint64(in string) uint64 {
	out, err := strconv.ParseUint(in, 10, 64)
	if err != nil {
		panic(fmt.Errorf("unable to parse %q as uint64: %w", in, err))
	}

	return out
}

func sanitizeBlockForCompare(block *pbbstream.Block) *pbbstream.Block {

	var antelopeBlock pbantelope.Block
	err := block.Payload.UnmarshalTo(&antelopeBlock)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal block payload into Antelope block: %w", err))
	}

	var sanitizeContext func(logContext *pbantelope.Exception_LogContext)
	sanitizeContext = func(logContext *pbantelope.Exception_LogContext) {
		if logContext != nil {
			logContext.Line = 0
			logContext.ThreadName = "thread"
			logContext.Timestamp = fixedTimestamp
			sanitizeContext(logContext.Context)
		}
	}

	sanitizeException := func(exception *pbantelope.Exception) {
		if exception != nil {
			for _, stack := range exception.Stack {
				sanitizeContext(stack.Context)
			}
		}
	}

	sanitizeRLimitOp := func(rlimitOp *pbantelope.RlimitOp) {
		switch v := rlimitOp.Kind.(type) {
		case *pbantelope.RlimitOp_AccountUsage:
			v.AccountUsage.CpuUsage.LastOrdinal = 111
			v.AccountUsage.NetUsage.LastOrdinal = 222
		case *pbantelope.RlimitOp_State:
			v.State.AverageBlockCpuUsage.LastOrdinal = 333
			v.State.AverageBlockNetUsage.LastOrdinal = 444
		}
	}

	for _, rlimitOp := range antelopeBlock.RlimitOps {
		sanitizeRLimitOp(rlimitOp)
	}

	for _, trxTrace := range antelopeBlock.UnfilteredTransactionTraces {
		trxTrace.Elapsed = 888
		sanitizeException(trxTrace.Exception)

		for _, permOp := range trxTrace.PermOps {
			if permOp.OldPerm != nil {
				permOp.OldPerm.LastUpdated = fixedTimestamp
			}

			if permOp.NewPerm != nil {
				permOp.NewPerm.LastUpdated = fixedTimestamp
			}
		}

		for _, rlimitOp := range trxTrace.RlimitOps {
			sanitizeRLimitOp(rlimitOp)
		}

		for _, actTrace := range trxTrace.ActionTraces {
			actTrace.Elapsed = 999
			sanitizeException(actTrace.Exception)
		}

		if trxTrace.FailedDtrxTrace != nil {
			trxTrace.FailedDtrxTrace.Elapsed = 101010
			sanitizeException(trxTrace.FailedDtrxTrace.Exception)
			for _, actTrace := range trxTrace.FailedDtrxTrace.ActionTraces {
				actTrace.Elapsed = 111111
				sanitizeException(actTrace.Exception)
			}
		}
	}

	res, err := firecore.EncodeBlock(&antelopeBlock)
	if err != nil {
		panic(fmt.Errorf("failed to encode Antelope block: %w", err))
	}

	return res
}
