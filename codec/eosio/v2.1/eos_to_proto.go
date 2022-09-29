package eosio

import (
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"math"
	"sort"
	"time"

	"github.com/EOS-Nation/firehose-antelope/codec/eosio"
	"github.com/EOS-Nation/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"go.uber.org/zap"
)

func TransactionReceiptToDEOS(txReceipt *TransactionReceipt) *pbantelope.TransactionReceipt {
	receipt := &pbantelope.TransactionReceipt{
		Status:               TransactionStatusToDEOS(txReceipt.Status),
		CpuUsageMicroSeconds: txReceipt.CPUUsageMicroSeconds,
		NetUsageWords:        uint32(txReceipt.NetUsageWords),
	}

	receipt.Id = txReceipt.Transaction.ID.String()
	if txReceipt.Transaction.Packed != nil {
		receipt.PackedTransaction = &pbantelope.PackedTransaction{
			Signatures:            SignaturesToDEOS(txReceipt.Transaction.Packed.Signatures),
			Compression:           uint32(txReceipt.Transaction.Packed.Compression),
			PackedContextFreeData: txReceipt.Transaction.Packed.PackedContextFreeData,
			PackedTransaction:     txReceipt.Transaction.Packed.PackedTransaction,
		}
	}

	return receipt
}

func PackedTransactionToDEOS(in *PackedTransaction) *pbantelope.PackedTransaction {
	out := &pbantelope.PackedTransaction{
		Compression:       uint32(in.Compression),
		PackedTransaction: in.PackedTransaction,
	}

	switch in.PrunableData.TypeID {
	case PrunableDataVariant.TypeID("full_legacy"):
		fullLegacy := in.PrunableData.Impl.(*PackedTransactionPrunableFullLegacy)

		out.Signatures = eosio.SignaturesToDEOS(fullLegacy.Signatures)
		out.PackedContextFreeData = fullLegacy.PackedContextFreeData

	case PrunableDataVariant.TypeID("full"):
		panic(fmt.Errorf("Only full_legacy pruning state is supported right now, got full"))
		// full := in.PrunableData.Impl.(*PackedTransactionPrunableFull)
		// out.Signatures = eosio.SignaturesToDEOS(full.Signatures)

	case PrunableDataVariant.TypeID("partial"):
		panic(fmt.Errorf("Only full_legacy pruning state is supported right now, got partial"))
		// partial := in.PrunableData.Impl.(*PackedTransactionPrunablePartial)
		// out.Signatures = eosio.SignaturesToDEOS(partial.Signatures)

	case PrunableDataVariant.TypeID("none"):
		panic(fmt.Errorf("Only full_legacy pruning state is supported right now, got none"))

	default:
		id, name, _ := in.PrunableData.Obtain(PrunableDataVariant)
		panic(fmt.Errorf("PrunableData variant %q (%d) is unknown", name, id))
	}

	return out
}

func TransactionTraceToDEOS(logger *zap.Logger, in *TransactionTrace, opts ...eosio.ConversionOption) *pbantelope.TransactionTrace {
	id := in.ID.String()

	out := &pbantelope.TransactionTrace{
		Id:              id,
		BlockNum:        uint64(in.BlockNum),
		BlockTime:       mustProtoTimestamp(in.BlockTime.Time),
		ProducerBlockId: in.ProducerBlockID.String(),
		Elapsed:         int64(in.Elapsed),
		NetUsage:        uint64(in.NetUsage),
		Scheduled:       in.Scheduled,
		Exception:       eosio.ExceptionToDEOS(in.Except),
		ErrorCode:       eosio.ErrorCodeToDEOS(in.ErrorCode),
	}

	var someConsoleTruncated bool
	out.ActionTraces, someConsoleTruncated = ActionTracesToDEOS(in.ActionTraces, opts...)
	if someConsoleTruncated {
		logger.Info("transaction had some of its action trace's console entries truncated", zap.String("id", id))
	}

	if in.FailedDtrxTrace != nil {
		out.FailedDtrxTrace = TransactionTraceToDEOS(logger, in.FailedDtrxTrace, opts...)
	}
	if in.Receipt != nil {
		out.Receipt = eosio.TransactionReceiptHeaderToDEOS(in.Receipt)
	}

	return out
}

func ActionTracesToDEOS(actionTraces []*ActionTrace, opts ...eosio.ConversionOption) (out []*pbantelope.ActionTrace, someConsoleTruncated bool) {
	if len(actionTraces) <= 0 {
		return nil, false
	}

	sort.Slice(actionTraces, func(i, j int) bool {
		leftSeq := uint64(math.MaxUint64)
		rightSeq := uint64(math.MaxUint64)

		if leftReceipt := actionTraces[i].Receipt; leftReceipt != nil {
			if seq := leftReceipt.GlobalSequence; seq != 0 {
				leftSeq = uint64(seq)
			}
		}
		if rightReceipt := actionTraces[j].Receipt; rightReceipt != nil {
			if seq := rightReceipt.GlobalSequence; seq != 0 {
				rightSeq = uint64(seq)
			}
		}

		return leftSeq < rightSeq
	})

	out = make([]*pbantelope.ActionTrace, len(actionTraces))
	var consoleTruncated bool
	for idx, actionTrace := range actionTraces {
		out[idx], consoleTruncated = ActionTraceToDEOS(actionTrace, uint32(idx), opts...)
		if consoleTruncated {
			someConsoleTruncated = true
		}
	}

	return
}

func ActionTraceToDEOS(in *ActionTrace, execIndex uint32, opts ...eosio.ConversionOption) (out *pbantelope.ActionTrace, consoleTruncated bool) {
	out = &pbantelope.ActionTrace{
		Receiver:         string(in.Receiver),
		Action:           eosio.ActionToDEOS(in.Action),
		Elapsed:          int64(in.ElapsedUs),
		Console:          string(in.Console),
		TransactionId:    in.TransactionID.String(),
		ContextFree:      in.ContextFree,
		ProducerBlockId:  in.ProducerBlockID.String(),
		BlockNum:         uint64(in.BlockNum),
		BlockTime:        mustProtoTimestamp(in.BlockTime.Time),
		AccountRamDeltas: AccountRAMDeltasToDEOS(in.AccountRAMDeltas),
		// AccountDiskDeltas:    AccountDeltasToDEOS(in.AccountDiskDeltas),
		Exception:            eosio.ExceptionToDEOS(in.Except),
		ActionOrdinal:        uint32(in.ActionOrdinal),
		CreatorActionOrdinal: uint32(in.CreatorActionOrdinal),
		ExecutionIndex:       execIndex,
		ErrorCode:            eosio.ErrorCodeToDEOS(in.ErrorCode),
		RawReturnValue:       in.ReturnValue,
	}
	out.ClosestUnnotifiedAncestorActionOrdinal = uint32(in.ClosestUnnotifiedAncestorActionOrdinal) // freaking long line, stay away from me

	if in.Receipt != nil {
		out.Receipt = eosio.ActionTraceReceiptToDEOS(in.Receipt)
	}

	initialConsoleLength := len(in.Console)
	for _, opt := range opts {
		if v, ok := opt.(eosio.ActionConversionOption); ok {
			v.Apply(out)
		}
	}

	return out, initialConsoleLength != len(out.Console)
}

func AccountDeltasToDEOS(deltas []AccountDelta) (out []*pbantelope.AccountDelta) {
	if len(deltas) <= 0 {
		return nil
	}

	out = make([]*pbantelope.AccountDelta, len(deltas))
	for i, delta := range deltas {
		out[i] = &pbantelope.AccountDelta{
			Account: string(delta.Account),
			Delta:   int64(delta.Delta),
		}
	}
	return
}

func AccountRAMDeltasToDEOS(deltas []AccountDelta) (out []*pbantelope.AccountRAMDelta) {
	if len(deltas) <= 0 {
		return nil
	}

	out = make([]*pbantelope.AccountRAMDelta, len(deltas))
	for i, delta := range deltas {
		out[i] = &pbantelope.AccountRAMDelta{
			Account: string(delta.Account),
			Delta:   int64(delta.Delta),
		}
	}
	return
}

func TransactionStatusToDEOS(in eos.TransactionStatus) pbantelope.TransactionStatus {
	switch in {
	case eos.TransactionStatusExecuted:
		return pbantelope.TransactionStatus_TRANSACTIONSTATUS_EXECUTED
	case eos.TransactionStatusSoftFail:
		return pbantelope.TransactionStatus_TRANSACTIONSTATUS_SOFTFAIL
	case eos.TransactionStatusHardFail:
		return pbantelope.TransactionStatus_TRANSACTIONSTATUS_HARDFAIL
	case eos.TransactionStatusDelayed:
		return pbantelope.TransactionStatus_TRANSACTIONSTATUS_DELAYED
	case eos.TransactionStatusExpired:
		return pbantelope.TransactionStatus_TRANSACTIONSTATUS_EXPIRED
	default:
		return pbantelope.TransactionStatus_TRANSACTIONSTATUS_UNKNOWN
	}
}

func SignaturesToDEOS(in []ecc.Signature) (out []string) {
	out = make([]string, len(in))
	for i, signature := range in {
		out[i] = signature.String()
	}
	return
}

func mustProtoTimestamp(in time.Time) *timestamp.Timestamp {
	out, err := ptypes.TimestampProto(in)
	if err != nil {
		panic(fmt.Sprintf("invalid timestamp conversion %q: %s", in, err))
	}
	return out
}
