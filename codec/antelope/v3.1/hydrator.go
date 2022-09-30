package antelope

import (
	"fmt"
	"github.com/EOS-Nation/firehose-antelope/codec/antelope"

	"github.com/EOS-Nation/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/eoscanada/eos-go"
	"go.uber.org/zap"
)

func NewHydrator(parentLogger *zap.Logger) *Hydrator {
	return &Hydrator{
		logger: parentLogger.With(zap.String("antelope", "3.1.x")),
	}
}

type Hydrator struct {
	logger *zap.Logger
}

func (h *Hydrator) HydrateBlock(block *pbantelope.Block, input []byte) error {
	h.logger.Debug("hydrating block from bytes")

	blockState := &BlockState{}
	err := unmarshalBinary(input, blockState)
	if err != nil {
		return fmt.Errorf("unmarshalling binary block state (3.1.x): %w", err)
	}

	signedBlock := blockState.SignedBlock

	block.Id = blockState.BlockID.String()
	block.Number = blockState.BlockNum
	// Version 1: Added the total counts (ExecutedInputActionCount, ExecutedTotalActionCount,
	// TransactionCount, TransactionTraceCount)
	block.Version = 1
	block.Header = antelope.BlockHeaderToDEOS(&signedBlock.BlockHeader)
	block.BlockExtensions = antelope.ExtensionsToDEOS(signedBlock.BlockExtensions)
	block.DposIrreversibleBlocknum = blockState.DPoSIrreversibleBlockNum
	block.DposProposedIrreversibleBlocknum = blockState.DPoSProposedIrreversibleBlockNum
	// block.Validated = blockState.Validated
	block.BlockrootMerkle = antelope.BlockrootMerkleToDEOS(blockState.BlockrootMerkle)
	block.ProducerToLastProduced = antelope.ProducerToLastProducedToDEOS(blockState.ProducerToLastProduced)
	block.ProducerToLastImpliedIrb = antelope.ProducerToLastImpliedIrbToDEOS(blockState.ProducerToLastImpliedIRB)
	block.ActivatedProtocolFeatures = antelope.ActivatedProtocolFeaturesToDEOS(blockState.ActivatedProtocolFeatures)
	block.ProducerSignature = signedBlock.ProducerSignature.String()

	block.ConfirmCount = make([]uint32, len(blockState.ConfirmCount))
	for i, count := range blockState.ConfirmCount {
		block.ConfirmCount[i] = uint32(count)
	}

	if blockState.PendingSchedule != nil {
		block.PendingSchedule = antelope.PendingScheduleToDEOS(blockState.PendingSchedule)
	}

	block.ValidBlockSigningAuthorityV2 = antelope.BlockSigningAuthorityToDEOS(blockState.ValidBlockSigningAuthorityV2)
	block.ActiveScheduleV2 = antelope.ProducerAuthorityScheduleToDEOS(blockState.ActiveSchedule)

	block.UnfilteredTransactionCount = uint32(len(signedBlock.Transactions))
	for idx, transaction := range signedBlock.Transactions {
		deosTransaction := TransactionReceiptToDEOS(transaction)
		deosTransaction.Index = uint64(idx)

		block.UnfilteredTransactions = append(block.UnfilteredTransactions, deosTransaction)
	}

	block.UnfilteredTransactionTraceCount = uint32(len(block.UnfilteredTransactionTraces))
	for idx, t := range block.UnfilteredTransactionTraces {
		t.Index = uint64(idx)
		t.BlockTime = block.Header.Timestamp
		t.ProducerBlockId = block.Id
		t.BlockNum = uint64(block.Number)

		for _, actionTrace := range t.ActionTraces {
			block.UnfilteredExecutedTotalActionCount++
			if actionTrace.IsInput() {
				block.UnfilteredExecutedInputActionCount++
			}
		}
	}

	return nil
}

func (h *Hydrator) DecodeTransactionTrace(input []byte, opts ...antelope.ConversionOption) (*pbantelope.TransactionTrace, error) {
	trxTrace := &TransactionTrace{}
	if err := unmarshalBinary(input, trxTrace); err != nil {
		return nil, fmt.Errorf("unmarshalling binary transaction trace: %w", err)
	}

	return TransactionTraceToDEOS(h.logger, trxTrace, opts...), nil
}

func unmarshalBinary(data []byte, v interface{}) error {
	decoder := eos.NewDecoder(data)
	decoder.DecodeActions(false)
	decoder.DecodeP2PMessage(false)

	return decoder.Decode(v)
}
