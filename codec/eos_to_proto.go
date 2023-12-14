// Copyright 2019 dfuse Platform Inc.
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

package codec

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
)

func bytesSlicesToHexBytes(in [][]byte) []eos.HexBytes {
	out := make([]eos.HexBytes, len(in))
	for i, s := range in {
		out[i] = s
	}
	return out
}

func BlockHeaderToEOS(in *pbantelope.BlockHeader) *eos.BlockHeader {
	stamp := in.Timestamp.AsTime()
	prev, _ := hex.DecodeString(in.Previous)
	out := &eos.BlockHeader{
		Timestamp:        eos.BlockTimestamp{Time: stamp},
		Producer:         eos.AccountName(in.Producer),
		Confirmed:        uint16(in.Confirmed),
		Previous:         prev,
		TransactionMRoot: in.TransactionMroot,
		ActionMRoot:      in.ActionMroot,
		ScheduleVersion:  in.ScheduleVersion,
		HeaderExtensions: ExtensionsToEOS(in.HeaderExtensions),
	}

	if in.NewProducersV1 != nil {
		out.NewProducersV1 = ProducerScheduleToEOS(in.NewProducersV1)
	}

	return out
}

func BlockSigningAuthorityToEOS(in *pbantelope.BlockSigningAuthority) *eos.BlockSigningAuthority {
	switch v := in.Variant.(type) {
	case *pbantelope.BlockSigningAuthority_V0:
		return &eos.BlockSigningAuthority{
			BaseVariant: eos.BaseVariant{
				TypeID: eos.BlockSigningAuthorityVariant.TypeID("block_signing_authority_v0"),
				Impl: eos.BlockSigningAuthorityV0{
					Threshold: v.V0.Threshold,
					Keys:      KeyWeightsPToEOS(v.V0.Keys),
				},
			},
		}
	default:
		panic(fmt.Errorf("unknown block signing authority variant %t", in.Variant))
	}
}

func ProducerScheduleToEOS(in *pbantelope.ProducerSchedule) *eos.ProducerSchedule {
	return &eos.ProducerSchedule{
		Version:   in.Version,
		Producers: ProducerKeysToEOS(in.Producers),
	}
}

func ProducerAuthorityScheduleToEOS(in *pbantelope.ProducerAuthoritySchedule) *eos.ProducerAuthoritySchedule {
	return &eos.ProducerAuthoritySchedule{
		Version:   in.Version,
		Producers: ProducerAuthoritiesToEOS(in.Producers),
	}
}

func ProducerKeysToEOS(in []*pbantelope.ProducerKey) (out []eos.ProducerKey) {
	out = make([]eos.ProducerKey, len(in))
	for i, producer := range in {
		// panic on error instead?
		key, _ := ecc.NewPublicKey(producer.BlockSigningKey)

		out[i] = eos.ProducerKey{
			AccountName:     eos.AccountName(producer.AccountName),
			BlockSigningKey: key,
		}
	}
	return
}

func PublicKeysToEOS(in []string) (out []*ecc.PublicKey) {
	if len(in) <= 0 {
		return nil
	}
	out = make([]*ecc.PublicKey, len(in))
	for i, inkey := range in {
		// panic on error instead?
		key, _ := ecc.NewPublicKey(inkey)

		out[i] = &key
	}
	return
}

func ExtensionsToEOS(in []*pbantelope.Extension) (out []*eos.Extension) {
	if len(in) <= 0 {
		return nil
	}

	out = make([]*eos.Extension, len(in))
	for i, extension := range in {
		out[i] = &eos.Extension{
			Type: uint16(extension.Type),
			Data: extension.Data,
		}
	}
	return
}

func ProducerAuthoritiesToEOS(producerAuthorities []*pbantelope.ProducerAuthority) (out []*eos.ProducerAuthority) {
	if len(producerAuthorities) <= 0 {
		return nil
	}

	out = make([]*eos.ProducerAuthority, len(producerAuthorities))
	for i, authority := range producerAuthorities {
		out[i] = &eos.ProducerAuthority{
			AccountName:           eos.AccountName(authority.AccountName),
			BlockSigningAuthority: BlockSigningAuthorityToEOS(authority.BlockSigningAuthority),
		}
	}
	return
}

func TransactionReceiptHeaderToEOS(in *pbantelope.TransactionReceiptHeader) *eos.TransactionReceiptHeader {
	return &eos.TransactionReceiptHeader{
		Status:               TransactionStatusToEOS(in.Status),
		CPUUsageMicroSeconds: in.CpuUsageMicroSeconds,
		NetUsageWords:        eos.Varuint32(in.NetUsageWords),
	}
}

func SignaturesToEOS(in []string) []ecc.Signature {
	out := make([]ecc.Signature, len(in))
	for i, signature := range in {
		sig, err := ecc.NewSignature(signature)
		if err != nil {
			panic(fmt.Sprintf("failed to read signature %q: %s", signature, err))
		}

		out[i] = sig
	}
	return out
}

func TransactionTraceToEOS(in *pbantelope.TransactionTrace) (out *eos.TransactionTrace) {
	out = &eos.TransactionTrace{
		ID:              ChecksumToEOS(in.Id),
		BlockNum:        uint32(in.BlockNum),
		BlockTime:       TimestampToBlockTimestamp(in.BlockTime),
		ProducerBlockID: ChecksumToEOS(in.ProducerBlockId),
		Elapsed:         eos.Int64(in.Elapsed),
		NetUsage:        eos.Uint64(in.NetUsage),
		Scheduled:       in.Scheduled,
		ActionTraces:    ActionTracesToEOS(in.ActionTraces),
		Except:          ExceptionToEOS(in.Exception),
		ErrorCode:       ErrorCodeToEOS(in.ErrorCode),
	}

	if in.FailedDtrxTrace != nil {
		out.FailedDtrxTrace = TransactionTraceToEOS(in.FailedDtrxTrace)
	}
	if in.Receipt != nil {
		out.Receipt = TransactionReceiptHeaderToEOS(in.Receipt)
	}

	return out
}

func AuthoritiesToEOS(authority *pbantelope.Authority) eos.Authority {
	return eos.Authority{
		Threshold: authority.Threshold,
		Keys:      KeyWeightsToEOS(authority.Keys),
		Accounts:  PermissionLevelWeightsToEOS(authority.Accounts),
		Waits:     WaitWeightsToEOS(authority.Waits),
	}
}

func WaitWeightsToEOS(waits []*pbantelope.WaitWeight) (out []eos.WaitWeight) {
	if len(waits) <= 0 {
		return nil
	}

	out = make([]eos.WaitWeight, len(waits))
	for i, o := range waits {
		out[i] = eos.WaitWeight{
			WaitSec: o.WaitSec,
			Weight:  uint16(o.Weight),
		}
	}
	return out
}

func PermissionLevelWeightsToEOS(weights []*pbantelope.PermissionLevelWeight) (out []eos.PermissionLevelWeight) {
	if len(weights) == 0 {
		return []eos.PermissionLevelWeight{}
	}

	out = make([]eos.PermissionLevelWeight, len(weights))
	for i, o := range weights {
		out[i] = eos.PermissionLevelWeight{
			Permission: PermissionLevelToEOS(o.Permission),
			Weight:     uint16(o.Weight),
		}
	}
	return
}

func PermissionLevelToEOS(perm *pbantelope.PermissionLevel) eos.PermissionLevel {
	return eos.PermissionLevel{
		Actor:      eos.AccountName(perm.Actor),
		Permission: eos.PermissionName(perm.Permission),
	}
}

func KeyWeightsToEOS(keys []*pbantelope.KeyWeight) (out []eos.KeyWeight) {
	if len(keys) <= 0 {
		return nil
	}

	out = make([]eos.KeyWeight, len(keys))
	for i, o := range keys {
		out[i] = eos.KeyWeight{
			PublicKey: ecc.MustNewPublicKey(o.PublicKey),
			Weight:    uint16(o.Weight),
		}
	}
	return

}

func KeyWeightsPToEOS(keys []*pbantelope.KeyWeight) (out []*eos.KeyWeight) {
	if len(keys) <= 0 {
		return nil
	}

	out = make([]*eos.KeyWeight, len(keys))
	for i, o := range keys {
		out[i] = &eos.KeyWeight{
			PublicKey: ecc.MustNewPublicKey(o.PublicKey),
			Weight:    uint16(o.Weight),
		}
	}
	return

}

func TransactionToEOS(trx *pbantelope.Transaction) *eos.Transaction {
	var contextFreeActions []*eos.Action
	if len(trx.ContextFreeActions) > 0 {
		contextFreeActions = make([]*eos.Action, len(trx.ContextFreeActions))
		for i, act := range trx.ContextFreeActions {
			contextFreeActions[i] = ActionToEOS(act)
		}
	}

	var actions []*eos.Action
	if len(trx.Actions) > 0 {
		actions = make([]*eos.Action, len(trx.Actions))
		for i, act := range trx.Actions {
			actions[i] = ActionToEOS(act)
		}
	}

	return &eos.Transaction{
		TransactionHeader:  *(TransactionHeaderToEOS(trx.Header)),
		ContextFreeActions: contextFreeActions,
		Actions:            actions,
		Extensions:         ExtensionsToEOS(trx.Extensions),
	}
}

func TransactionHeaderToEOS(trx *pbantelope.TransactionHeader) *eos.TransactionHeader {
	out := &eos.TransactionHeader{
		Expiration:       TimestampToJSONTime(trx.Expiration),
		RefBlockNum:      uint16(trx.RefBlockNum),
		RefBlockPrefix:   uint32(trx.RefBlockPrefix),
		MaxNetUsageWords: eos.Varuint32(trx.MaxNetUsageWords),
		MaxCPUUsageMS:    uint8(trx.MaxCpuUsageMs),
		DelaySec:         eos.Varuint32(trx.DelaySec),
	}

	return out
}

func SignedTransactionToEOS(trx *pbantelope.SignedTransaction) *eos.SignedTransaction {
	return &eos.SignedTransaction{
		Transaction:     TransactionToEOS(trx.Transaction),
		Signatures:      SignaturesToEOS(trx.Signatures),
		ContextFreeData: bytesSlicesToHexBytes(trx.ContextFreeData),
	}
}

func ActionTracesToEOS(actionTraces []*pbantelope.ActionTrace) (out []eos.ActionTrace) {
	if len(actionTraces) <= 0 {
		return nil
	}

	out = make([]eos.ActionTrace, len(actionTraces))
	for i, actionTrace := range actionTraces {
		out[i] = ActionTraceToEOS(actionTrace)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ActionOrdinal < out[j].ActionOrdinal })

	return
}

func AuthSequenceListToEOS(in []*pbantelope.AuthSequence) (out []eos.TransactionTraceAuthSequence) {
	if len(in) == 0 {
		return []eos.TransactionTraceAuthSequence{}
	}

	out = make([]eos.TransactionTraceAuthSequence, len(in))
	for i, seq := range in {
		out[i] = AuthSequenceToEOS(seq)
	}

	return
}

func AuthSequenceToEOS(in *pbantelope.AuthSequence) eos.TransactionTraceAuthSequence {
	return eos.TransactionTraceAuthSequence{
		Account:  eos.AccountName(in.AccountName),
		Sequence: eos.Uint64(in.Sequence),
	}
}

func ErrorCodeToEOS(in uint64) *eos.Uint64 {
	if in != 0 {
		val := eos.Uint64(in)
		return &val
	}
	return nil
}

func ActionTraceToEOS(in *pbantelope.ActionTrace) (out eos.ActionTrace) {
	out = eos.ActionTrace{
		Receiver:             eos.AccountName(in.Receiver),
		Action:               ActionToEOS(in.Action),
		Elapsed:              eos.Int64(in.Elapsed),
		Console:              eos.SafeString(in.Console),
		TransactionID:        ChecksumToEOS(in.TransactionId),
		ContextFree:          in.ContextFree,
		ProducerBlockID:      ChecksumToEOS(in.ProducerBlockId),
		BlockNum:             uint32(in.BlockNum),
		BlockTime:            TimestampToBlockTimestamp(in.BlockTime),
		AccountRAMDeltas:     AccountRAMDeltasToEOS(in.AccountRamDeltas),
		Except:               ExceptionToEOS(in.Exception),
		ActionOrdinal:        eos.Varuint32(in.ActionOrdinal),
		CreatorActionOrdinal: eos.Varuint32(in.CreatorActionOrdinal),
		ErrorCode:            ErrorCodeToEOS(in.ErrorCode),
	}
	out.ClosestUnnotifiedAncestorActionOrdinal = eos.Varuint32(in.ClosestUnnotifiedAncestorActionOrdinal) // freaking long line, stay away from me

	if in.Receipt != nil {
		receipt := in.Receipt

		out.Receipt = &eos.ActionTraceReceipt{
			Receiver:        eos.AccountName(receipt.Receiver),
			ActionDigest:    ChecksumToEOS(receipt.Digest),
			GlobalSequence:  eos.Uint64(receipt.GlobalSequence),
			AuthSequence:    AuthSequenceListToEOS(receipt.AuthSequence),
			ReceiveSequence: eos.Uint64(receipt.RecvSequence),
			CodeSequence:    eos.Varuint32(receipt.CodeSequence),
			ABISequence:     eos.Varuint32(receipt.AbiSequence),
		}
	}

	return
}

func ChecksumToEOS(in string) eos.Checksum256 {
	out, err := hex.DecodeString(in)
	if err != nil {
		panic(fmt.Sprintf("failed decoding checksum %q: %s", in, err))
	}

	return eos.Checksum256(out)
}

func ActionToEOS(action *pbantelope.Action) (out *eos.Action) {
	d := eos.ActionData{}
	d.SetToServer(false) // rather, what we expect FROM `nodeos` servers

	d.HexData = eos.HexBytes(action.RawData)
	if len(action.JsonData) != 0 {
		err := json.Unmarshal([]byte(action.JsonData), &d.Data)
		if err != nil {
			panic(fmt.Sprintf("unmarshaling action json data %q: %s", action.JsonData, err))
		}
	}

	out = &eos.Action{
		Account:       eos.AccountName(action.Account),
		Name:          eos.ActionName(action.Name),
		Authorization: AuthorizationToEOS(action.Authorization),
		ActionData:    d,
	}

	return out
}

func AuthorizationToEOS(authorization []*pbantelope.PermissionLevel) (out []eos.PermissionLevel) {
	if len(authorization) == 0 {
		return []eos.PermissionLevel{}
	}

	out = make([]eos.PermissionLevel, len(authorization))
	for i, permission := range authorization {
		out[i] = PermissionLevelToEOS(permission)
	}
	return
}

func AccountRAMDeltasToEOS(deltas []*pbantelope.AccountRAMDelta) (out []*eos.AccountRAMDelta) {
	if len(deltas) == 0 {
		return []*eos.AccountRAMDelta{}
	}

	out = make([]*eos.AccountRAMDelta, len(deltas))
	for i, delta := range deltas {
		out[i] = &eos.AccountRAMDelta{
			Account: eos.AccountName(delta.Account),
			Delta:   eos.Int64(delta.Delta),
		}
	}
	return
}

func ExceptionToEOS(in *pbantelope.Exception) *eos.Except {
	if in == nil {
		return nil
	}
	out := &eos.Except{
		Code:    eos.Int64(in.Code),
		Name:    in.Name,
		Message: in.Message,
	}

	if len(in.Stack) > 0 {
		out.Stack = make([]*eos.ExceptLogMessage, len(in.Stack))
		for i, el := range in.Stack {
			msg := &eos.ExceptLogMessage{
				Format: el.Format,
			}

			ctx := LogContextToEOS(el.Context)
			if ctx != nil {
				msg.Context = *ctx
			}

			if len(el.Data) > 0 {
				msg.Data = json.RawMessage(el.Data)
			}

			out.Stack[i] = msg
		}
	}

	return out
}

func LogContextToEOS(in *pbantelope.Exception_LogContext) *eos.ExceptLogContext {
	if in == nil {
		return nil
	}

	var exceptLevel eos.ExceptLogLevel
	exceptLevel.FromString(in.Level)

	return &eos.ExceptLogContext{
		Level:      exceptLevel,
		File:       in.File,
		Line:       uint64(in.Line),
		Method:     in.Method,
		Hostname:   in.Hostname,
		ThreadName: in.ThreadName,
		Timestamp:  TimestampToJSONTime(in.Timestamp),
		Context:    LogContextToEOS(in.Context),
	}
}

func TimestampToJSONTime(in *timestamppb.Timestamp) eos.JSONTime {
	return eos.JSONTime{Time: in.AsTime()}
}

func TimestampToBlockTimestamp(in *timestamppb.Timestamp) eos.BlockTimestamp {
	return eos.BlockTimestamp{Time: in.AsTime()}
}

func TransactionStatusToEOS(in pbantelope.TransactionStatus) eos.TransactionStatus {
	switch in {
	case pbantelope.TransactionStatus_TRANSACTIONSTATUS_EXECUTED:
		return eos.TransactionStatusExecuted
	case pbantelope.TransactionStatus_TRANSACTIONSTATUS_SOFTFAIL:
		return eos.TransactionStatusSoftFail
	case pbantelope.TransactionStatus_TRANSACTIONSTATUS_HARDFAIL:
		return eos.TransactionStatusHardFail
	case pbantelope.TransactionStatus_TRANSACTIONSTATUS_DELAYED:
		return eos.TransactionStatusDelayed
	case pbantelope.TransactionStatus_TRANSACTIONSTATUS_EXPIRED:
		return eos.TransactionStatusExpired
	default:
		return eos.TransactionStatusUnknown
	}
}

func ExtractEOSSignedTransactionFromReceipt(trxReceipt *pbantelope.TransactionReceipt) (*eos.SignedTransaction, error) {
	eosPackedTx, err := pbantelopePackedTransactionToEOS(trxReceipt.PackedTransaction)
	if err != nil {
		return nil, fmt.Errorf("pbantelope.PackedTransaction to EOS conversion failed: %s", err)
	}

	signedTransaction, err := eosPackedTx.UnpackBare()
	if err != nil {
		return nil, fmt.Errorf("unable to unpack packed transaction: %s", err)
	}

	return signedTransaction, nil
}

// todo legacy code should not error anymore using timestamppb
//func mustProtoTimestamp(in time.Time) *timestamp.Timestamp {
//	out, err := ptypes.TimestampProto(in)
//	if err != nil {
//		panic(fmt.Sprintf("invalid timestamp conversion %q: %s", in, err))
//	}
//	return out
//}

func pbantelopePackedTransactionToEOS(packedTrx *pbantelope.PackedTransaction) (*eos.PackedTransaction, error) {
	signatures := make([]ecc.Signature, len(packedTrx.Signatures))
	for i, signature := range packedTrx.Signatures {
		eccSignature, err := ecc.NewSignature(signature)
		if err != nil {
			return nil, err
		}

		signatures[i] = eccSignature
	}

	return &eos.PackedTransaction{
		Signatures:            signatures,
		Compression:           eos.CompressionType(packedTrx.Compression),
		PackedContextFreeData: packedTrx.PackedContextFreeData,
		PackedTransaction:     packedTrx.PackedTransaction,
	}, nil
}
