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

package antelope

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ActivatedProtocolFeaturesToDEOS(in *eos.ProtocolFeatureActivationSet) *pbantelope.ActivatedProtocolFeatures {
	out := &pbantelope.ActivatedProtocolFeatures{}
	out.ProtocolFeatures = checksumsToBytesSlices(in.ProtocolFeatures)
	return out
}

func PendingScheduleToDEOS(in *eos.PendingSchedule) *pbantelope.PendingProducerSchedule {
	out := &pbantelope.PendingProducerSchedule{
		ScheduleLibNum: in.ScheduleLIBNum,
		ScheduleHash:   []byte(in.ScheduleHash),
	}

	/// Specific versions handling

	// Only in EOSIO 1.x
	if in.Schedule.V1 != nil {
		out.ScheduleV1 = ProducerScheduleToDEOS(in.Schedule.V1)
	}

	// Only in EOSIO 2.x
	if in.Schedule.V2 != nil {
		out.ScheduleV2 = ProducerAuthorityScheduleToDEOS(in.Schedule.V2)
	}

	// End (versions)

	return out
}

func ProducerToLastProducedToDEOS(in []eos.PairAccountNameBlockNum) []*pbantelope.ProducerToLastProduced {
	out := make([]*pbantelope.ProducerToLastProduced, len(in))
	for i, elem := range in {
		out[i] = &pbantelope.ProducerToLastProduced{
			Name:                 string(elem.AccountName),
			LastBlockNumProduced: uint32(elem.BlockNum),
		}
	}
	return out
}

func ProducerToLastImpliedIrbToDEOS(in []eos.PairAccountNameBlockNum) []*pbantelope.ProducerToLastImpliedIRB {
	out := make([]*pbantelope.ProducerToLastImpliedIRB, len(in))
	for i, elem := range in {
		out[i] = &pbantelope.ProducerToLastImpliedIRB{
			Name:                 string(elem.AccountName),
			LastBlockNumProduced: uint32(elem.BlockNum),
		}
	}
	return out
}

func BlockrootMerkleToDEOS(merkle *eos.MerkleRoot) *pbantelope.BlockRootMerkle {
	return &pbantelope.BlockRootMerkle{
		NodeCount:   uint32(merkle.NodeCount),
		ActiveNodes: checksumsToBytesSlices(merkle.ActiveNodes),
	}
}

func BlockHeaderToDEOS(blockHeader *eos.BlockHeader) *pbantelope.BlockHeader {
	out := &pbantelope.BlockHeader{
		Timestamp:        timestamppb.New(blockHeader.Timestamp.Time),
		Producer:         string(blockHeader.Producer),
		Confirmed:        uint32(blockHeader.Confirmed),
		Previous:         blockHeader.Previous.String(),
		TransactionMroot: blockHeader.TransactionMRoot,
		ActionMroot:      blockHeader.ActionMRoot,
		ScheduleVersion:  blockHeader.ScheduleVersion,
		HeaderExtensions: ExtensionsToDEOS(blockHeader.HeaderExtensions),
	}

	if blockHeader.NewProducersV1 != nil {
		out.NewProducersV1 = ProducerScheduleToDEOS(blockHeader.NewProducersV1)
	}

	return out
}

func BlockSigningAuthorityToDEOS(authority *eos.BlockSigningAuthority) *pbantelope.BlockSigningAuthority {
	out := &pbantelope.BlockSigningAuthority{}

	switch v := authority.Impl.(type) {
	case *eos.BlockSigningAuthorityV0:
		out.Variant = &pbantelope.BlockSigningAuthority_V0{
			V0: &pbantelope.BlockSigningAuthorityV0{
				Threshold: v.Threshold,
				Keys:      KeyWeightsPToDEOS(v.Keys),
			},
		}
	default:
		panic(fmt.Errorf("unable to convert eos.BlockSigningAuthority to deos: wrong type %T", authority.Impl))
	}

	return out
}

func ProducerScheduleToDEOS(e *eos.ProducerSchedule) *pbantelope.ProducerSchedule {
	return &pbantelope.ProducerSchedule{
		Version:   uint32(e.Version),
		Producers: ProducerKeysToDEOS(e.Producers),
	}
}

func ProducerAuthorityScheduleToDEOS(e *eos.ProducerAuthoritySchedule) *pbantelope.ProducerAuthoritySchedule {
	return &pbantelope.ProducerAuthoritySchedule{
		Version:   uint32(e.Version),
		Producers: ProducerAuthoritiesToDEOS(e.Producers),
	}
}

func ProducerKeysToDEOS(in []eos.ProducerKey) (out []*pbantelope.ProducerKey) {
	out = make([]*pbantelope.ProducerKey, len(in))
	for i, key := range in {
		out[i] = &pbantelope.ProducerKey{
			AccountName:     string(key.AccountName),
			BlockSigningKey: key.BlockSigningKey.String(),
		}
	}
	return
}

func ExtensionsToDEOS(in []*eos.Extension) (out []*pbantelope.Extension) {
	out = make([]*pbantelope.Extension, len(in))
	for i, extension := range in {
		out[i] = &pbantelope.Extension{
			Type: uint32(extension.Type),
			Data: extension.Data,
		}
	}

	return
}

func ProducerAuthoritiesToDEOS(producerAuthorities []*eos.ProducerAuthority) (out []*pbantelope.ProducerAuthority) {
	if len(producerAuthorities) <= 0 {
		return nil
	}

	out = make([]*pbantelope.ProducerAuthority, len(producerAuthorities))
	for i, authority := range producerAuthorities {
		out[i] = &pbantelope.ProducerAuthority{
			AccountName:           string(authority.AccountName),
			BlockSigningAuthority: BlockSigningAuthorityToDEOS(authority.BlockSigningAuthority),
		}
	}
	return
}

func TransactionReceiptToDEOS(txReceipt *eos.TransactionReceipt) *pbantelope.TransactionReceipt {
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

func TransactionReceiptHeaderToDEOS(in *eos.TransactionReceiptHeader) *pbantelope.TransactionReceiptHeader {
	return &pbantelope.TransactionReceiptHeader{
		Status:               TransactionStatusToDEOS(in.Status),
		CpuUsageMicroSeconds: in.CPUUsageMicroSeconds,
		NetUsageWords:        uint32(in.NetUsageWords),
	}
}

func SignaturesToDEOS(in []ecc.Signature) (out []string) {
	out = make([]string, len(in))
	for i, signature := range in {
		out[i] = signature.String()
	}
	return
}

func PermissionToDEOS(perm *eos.Permission) *pbantelope.Permission {
	return &pbantelope.Permission{
		Name:         perm.PermName,
		Parent:       perm.Parent,
		RequiredAuth: AuthoritiesToDEOS(&perm.RequiredAuth),
	}
}

func AuthoritiesToDEOS(authority *eos.Authority) *pbantelope.Authority {
	return &pbantelope.Authority{
		Threshold: authority.Threshold,
		Keys:      KeyWeightsToDEOS(authority.Keys),
		Accounts:  PermissionLevelWeightsToDEOS(authority.Accounts),
		Waits:     WaitWeightsToDEOS(authority.Waits),
	}
}

func WaitWeightsToDEOS(waits []eos.WaitWeight) (out []*pbantelope.WaitWeight) {
	if len(waits) <= 0 {
		return nil
	}

	out = make([]*pbantelope.WaitWeight, len(waits))
	for i, o := range waits {
		out[i] = &pbantelope.WaitWeight{
			WaitSec: o.WaitSec,
			Weight:  uint32(o.Weight),
		}
	}
	return out
}

func PermissionLevelWeightsToDEOS(weights []eos.PermissionLevelWeight) (out []*pbantelope.PermissionLevelWeight) {
	if len(weights) <= 0 {
		return nil
	}

	out = make([]*pbantelope.PermissionLevelWeight, len(weights))
	for i, o := range weights {
		out[i] = &pbantelope.PermissionLevelWeight{
			Permission: PermissionLevelToDEOS(o.Permission),
			Weight:     uint32(o.Weight),
		}
	}
	return
}

func PermissionLevelToDEOS(perm eos.PermissionLevel) *pbantelope.PermissionLevel {
	return &pbantelope.PermissionLevel{
		Actor:      string(perm.Actor),
		Permission: string(perm.Permission),
	}
}

func KeyWeightsToDEOS(keys []eos.KeyWeight) (out []*pbantelope.KeyWeight) {
	if len(keys) <= 0 {
		return nil
	}

	out = make([]*pbantelope.KeyWeight, len(keys))
	for i, o := range keys {
		out[i] = &pbantelope.KeyWeight{
			PublicKey: o.PublicKey.String(),
			Weight:    uint32(o.Weight),
		}
	}
	return
}

func KeyWeightsPToDEOS(keys []*eos.KeyWeight) (out []*pbantelope.KeyWeight) {
	if len(keys) <= 0 {
		return nil
	}

	out = make([]*pbantelope.KeyWeight, len(keys))
	for i, o := range keys {
		out[i] = &pbantelope.KeyWeight{
			PublicKey: o.PublicKey.String(),
			Weight:    uint32(o.Weight),
		}
	}
	return
}

func TransactionToDEOS(trx *eos.Transaction) *pbantelope.Transaction {
	var contextFreeActions []*pbantelope.Action
	if len(trx.ContextFreeActions) > 0 {
		contextFreeActions = make([]*pbantelope.Action, len(trx.ContextFreeActions))
		for i, act := range trx.ContextFreeActions {
			contextFreeActions[i] = ActionToDEOS(act)
		}
	}

	var actions []*pbantelope.Action
	if len(trx.Actions) > 0 {
		actions = make([]*pbantelope.Action, len(trx.Actions))
		for i, act := range trx.Actions {
			actions[i] = ActionToDEOS(act)
		}
	}

	return &pbantelope.Transaction{
		Header:             TransactionHeaderToDEOS(&trx.TransactionHeader),
		ContextFreeActions: contextFreeActions,
		Actions:            actions,
		Extensions:         ExtensionsToDEOS(trx.Extensions),
	}
}

func TransactionHeaderToDEOS(trx *eos.TransactionHeader) *pbantelope.TransactionHeader {
	out := &pbantelope.TransactionHeader{
		Expiration:       timestamppb.New(trx.Expiration.Time),
		RefBlockNum:      uint32(trx.RefBlockNum),
		RefBlockPrefix:   trx.RefBlockPrefix,
		MaxNetUsageWords: uint32(trx.MaxNetUsageWords),
		MaxCpuUsageMs:    uint32(trx.MaxCPUUsageMS),
		DelaySec:         uint32(trx.DelaySec),
	}

	return out
}

func SignedTransactionToDEOS(trx *eos.SignedTransaction) *pbantelope.SignedTransaction {
	return &pbantelope.SignedTransaction{
		Transaction:     TransactionToDEOS(trx.Transaction),
		Signatures:      SignaturesToDEOS(trx.Signatures),
		ContextFreeData: hexBytesToBytesSlices(trx.ContextFreeData),
	}
}

func CreationTreeToDEOS(tree CreationFlatTree) []*pbantelope.CreationFlatNode {
	if len(tree) <= 0 {
		return nil
	}

	out := make([]*pbantelope.CreationFlatNode, len(tree))
	for i, node := range tree {
		out[i] = &pbantelope.CreationFlatNode{
			CreatorActionIndex:   int32(node[1]),
			ExecutionActionIndex: uint32(node[2]),
		}
	}
	return out
}

func AuthSequenceToDEOS(in eos.TransactionTraceAuthSequence) *pbantelope.AuthSequence {
	return &pbantelope.AuthSequence{
		AccountName: string(in.Account),
		Sequence:    uint64(in.Sequence),
	}
}

func ActionTraceReceiptToDEOS(in *eos.ActionTraceReceipt) *pbantelope.ActionReceipt {
	authSequences := in.AuthSequence

	var deosAuthSequence []*pbantelope.AuthSequence
	if len(authSequences) > 0 {
		deosAuthSequence = make([]*pbantelope.AuthSequence, len(authSequences))
		for i, seq := range authSequences {
			deosAuthSequence[i] = AuthSequenceToDEOS(seq)
		}
	}

	return &pbantelope.ActionReceipt{
		Receiver:       string(in.Receiver),
		Digest:         in.ActionDigest.String(),
		GlobalSequence: uint64(in.GlobalSequence),
		AuthSequence:   deosAuthSequence,
		RecvSequence:   uint64(in.ReceiveSequence),
		CodeSequence:   uint64(in.CodeSequence),
		AbiSequence:    uint64(in.ABISequence),
	}
}

func ErrorCodeToDEOS(in *eos.Uint64) uint64 {
	if in != nil {
		return uint64(*in)
	}
	return 0
}

func ActionToDEOS(action *eos.Action) *pbantelope.Action {
	deosAction := &pbantelope.Action{
		Account:       string(action.Account),
		Name:          string(action.Name),
		Authorization: AuthorizationToDEOS(action.Authorization),
		RawData:       action.HexData,
	}

	if action.Data != nil {
		v, dataIsString := action.Data.(string)
		if dataIsString && len(action.HexData) == 0 {
			// When the action.Data is actually a string, and the HexData field is not set, we assume data sould be rawData instead
			rawData, err := hex.DecodeString(v)
			if err != nil {
				panic(fmt.Errorf("unable to unmarshal action data %q as hex: %s", v, err))
			}

			deosAction.RawData = rawData
		} else {
			serializedData, err := json.Marshal(action.Data)
			if err != nil {
				panic(fmt.Errorf("unable to unmarshal action data JSON: %s", err))
			}

			deosAction.JsonData = string(serializedData)
		}
	}

	return deosAction
}

func AuthorizationToDEOS(authorization []eos.PermissionLevel) (out []*pbantelope.PermissionLevel) {
	if len(authorization) <= 0 {
		return nil
	}

	out = make([]*pbantelope.PermissionLevel, len(authorization))
	for i, permission := range authorization {
		out[i] = PermissionLevelToDEOS(permission)
	}
	return
}

func AccountRAMDeltasToDEOS(deltas []*eos.AccountRAMDelta) (out []*pbantelope.AccountRAMDelta) {
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

func ExceptionToDEOS(in *eos.Except) *pbantelope.Exception {
	if in == nil {
		return nil
	}
	out := &pbantelope.Exception{
		Code:    int32(in.Code),
		Name:    in.Name,
		Message: in.Message,
	}

	if len(in.Stack) > 0 {
		out.Stack = make([]*pbantelope.Exception_LogMessage, len(in.Stack))
		for i, el := range in.Stack {
			out.Stack[i] = &pbantelope.Exception_LogMessage{
				Context: LogContextToDEOS(el.Context),
				Format:  el.Format,
				Data:    el.Data,
			}
		}
	}

	return out
}

func LogContextToDEOS(in eos.ExceptLogContext) *pbantelope.Exception_LogContext {
	out := &pbantelope.Exception_LogContext{
		Level:      in.Level.String(),
		File:       in.File,
		Line:       int32(in.Line),
		Method:     in.Method,
		Hostname:   in.Hostname,
		ThreadName: in.ThreadName,
		Timestamp:  timestamppb.New(in.Timestamp.Time),
	}
	if in.Context != nil {
		out.Context = LogContextToDEOS(*in.Context)
	}
	return out
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

func checksumsToBytesSlices(in []eos.Checksum256) [][]byte {
	out := make([][]byte, len(in))
	for i, s := range in {
		out[i] = s
	}
	return out
}

func hexBytesToBytesSlices(in []eos.HexBytes) [][]byte {
	out := make([][]byte, len(in))
	for i, s := range in {
		out[i] = s
	}
	return out
}
