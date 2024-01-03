package main

import (
	"fmt"
	pbantelope "github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"time"
)

func mustParseUint64(in string) uint64 {
	out, err := strconv.ParseUint(in, 10, 64)
	if err != nil {
		panic(fmt.Errorf("unable to parse %q as uint64: %w", in, err))
	}

	return out
}

func sanitizeBlockForCompare(block *pbantelope.Block) *pbantelope.Block {

	fixedTimestamp := timestamppb.New(time.Now())

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

	for _, rlimitOp := range block.RlimitOps {
		sanitizeRLimitOp(rlimitOp)
	}

	for _, trxTrace := range block.UnfilteredTransactionTraces {
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
			sanitizeException(trxTrace.FailedDtrxTrace.Exception)
			for _, actTrace := range trxTrace.FailedDtrxTrace.ActionTraces {
				sanitizeException(actTrace.Exception)
			}
		}
	}

	return block
}
