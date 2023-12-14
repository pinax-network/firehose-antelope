package ct

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pinax-network/firehose-antelope/types"
	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/bstream/pb/sf/bstream/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
	"github.com/mitchellh/go-testing-interface"
	"github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/streamingfast/jsonpb"
	"github.com/streamingfast/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type Hash string
type BlockTime string
type BlockTimestamp time.Time

var zlog *zap.Logger

func init() {
	logging.Register("github.com/dfuse-io/dfuse-eosio/codec/testing", &zlog)
}

func (h Hash) Bytes(t testing.T) []byte {
	bytes, err := hex.DecodeString(string(h))
	require.NoErrorf(t, err, "hash %q is to valid hex: %w", h)

	return bytes
}

func AutoGlobalSequence() *autoGlobalSequence {
	return &autoGlobalSequence{
		count: atomic.NewUint64(0),
	}
}

type autoGlobalSequence struct {
	count *atomic.Uint64
}

type FilteredBlock struct {
	Include         string
	Exclude         string
	UnfilteredStats Counts
	FilteredStats   Counts
}

type Counts struct {
	TrxTraceCount      int
	ActTraceInputCount int
	ActTraceTotalCount int
}

func Block(t testing.T, blkID string, components ...interface{}) *pbantelope.Block {
	ref := bstream.NewBlockRefFromID(blkID)

	pbblock := &pbantelope.Block{
		Id:     blkID,
		Number: uint32(ref.Num()),
	}

	blockTime, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05.5Z")
	require.NoError(t, err)

	pbblock.DposIrreversibleBlocknum = pbblock.Number - 1
	pbblock.Header = &pbantelope.BlockHeader{
		Previous:  fmt.Sprintf("%08x%s", pbblock.Number-1, blkID[8:]),
		Producer:  "tester",
		Timestamp: timestamppb.New(blockTime),
	}

	for _, component := range components {
		switch v := component.(type) {
		case BlockTime:
			blockTime, err := time.Parse(time.RFC3339, string(v))
			require.NoError(t, err)

			pbblock.Header.Timestamp = timestamppb.New(blockTime)
		case BlockTimestamp:
			pbblock.Header.Timestamp = timestamppb.New(time.Time(v))

		case *pbantelope.TransactionTrace:
			v.BlockNum = pbblock.Num()
			v.BlockTime = timestamppb.New(pbblock.Time())

			pbblock.UnfilteredTransactionTraces = append(pbblock.UnfilteredTransactionTraces, v)
		case *pbantelope.TrxOp:
			pbblock.UnfilteredImplicitTransactionOps = append(pbblock.UnfilteredImplicitTransactionOps, v)
		case *autoGlobalSequence:
		case FilteredBlock:
			// Performed at the very end
		default:
			failInvalidComponent(t, "block", component)
		}
	}

	pbblock.MigrateV0ToV1()

	if component := findTypedComponent(components, (*autoGlobalSequence)(nil)); component != nil {
		globalSequence := component.(*autoGlobalSequence)
		for _, trxTrace := range pbblock.UnfilteredTransactionTraces {
			for _, actTrace := range trxTrace.ActionTraces {
				// We only deal with set Receipt, if it's not set, we assume the caller wanted it like it
				if actTrace.Receipt != nil {
					sequence := globalSequence.count.Inc()
					actTrace.Receipt.GlobalSequence = sequence
				}
			}
		}
	}

	// Need to go at the end to ensure we catch all transaction traces
	if component := findComponent(components, func(component interface{}) bool { _, ok := component.(FilteredBlock); return ok }); component != nil {
		filtered := component.(FilteredBlock)

		pbblock.FilteringApplied = true
		pbblock.FilteringIncludeFilterExpr = filtered.Include
		pbblock.FilteringExcludeFilterExpr = filtered.Exclude
		pbblock.FilteredTransactionTraces = pbblock.UnfilteredTransactionTraces
		pbblock.UnfilteredTransactionTraces = nil
		pbblock.MigrateV0ToV1()

		pbblock.UnfilteredTransactionTraceCount = uint32(filtered.UnfilteredStats.TrxTraceCount)
		pbblock.UnfilteredExecutedInputActionCount = uint32(filtered.UnfilteredStats.ActTraceInputCount)
		pbblock.UnfilteredExecutedTotalActionCount = uint32(filtered.UnfilteredStats.ActTraceTotalCount)

		pbblock.FilteredTransactionTraceCount = uint32(filtered.FilteredStats.TrxTraceCount)
		pbblock.FilteredExecutedInputActionCount = uint32(filtered.FilteredStats.ActTraceInputCount)
		pbblock.FilteredExecutedTotalActionCount = uint32(filtered.FilteredStats.ActTraceTotalCount)
	}

	if os.Getenv("DEBUG") != "" {
		marshaler := &jsonpb.Marshaler{}
		out, err := marshaler.MarshalToString(pbblock)
		require.NoError(t, err)

		// We re-normalize to a plain map[string]interface{} so it's printed as JSON and not a proto default String implementation
		normalizedOut := map[string]interface{}{}
		require.NoError(t, json.Unmarshal([]byte(out), &normalizedOut))

		zlog.Debug("created test block", zap.Any("block", normalizedOut))
	}

	return pbblock
}

func ToBstreamBlock(t testing.T, block *pbantelope.Block) *pbbstream.Block {
	blk, err := types.BlockFromProto(block, block.LIBNum())
	require.NoError(t, err)

	return blk
}

func ToBstreamBlocks(t testing.T, blocks []*pbantelope.Block) (out []*pbbstream.Block) {
	out = make([]*pbbstream.Block, len(blocks))
	for i, block := range blocks {
		out[i] = ToBstreamBlock(t, block)
	}
	return
}

type TrxID string

func TrxTrace(t testing.T, components ...interface{}) *pbantelope.TransactionTrace {
	trace := &pbantelope.TransactionTrace{
		Receipt: &pbantelope.TransactionReceiptHeader{
			Status: pbantelope.TransactionStatus_TRANSACTIONSTATUS_EXECUTED,
		},
	}

	for _, element := range components {
		switch v := element.(type) {
		case TrxID:
			trace.Id = string(v)
		case *pbantelope.ActionTrace:
			trace.ActionTraces = append(trace.ActionTraces, v)
		case *pbantelope.DBOp:
			trace.DbOps = append(trace.DbOps, v)
		case *pbantelope.DTrxOp:
			trace.DtrxOps = append(trace.DtrxOps, v)
		case *pbantelope.TableOp:
			trace.TableOps = append(trace.TableOps, v)
		case pbantelope.TransactionStatus:
			trace.Receipt.Status = v
		default:
			failInvalidComponent(t, "transaction trace", element)
		}
	}

	for i, actTrace := range trace.ActionTraces {
		// Let's auto-assign all ExecutionIndex automatically if they are unset
		if actTrace.ExecutionIndex == 0 {
			// Unless it's 0 and not the first one, in which case someone already played with it and we should not change it
			if i != 0 {
				actTrace.ExecutionIndex = uint32(i)
			}
		}
	}

	return trace
}

func SignedTrx(t testing.T, elements ...interface{}) *pbantelope.SignedTransaction {
	signedTrx := &pbantelope.SignedTransaction{}
	signedTrx.Transaction = Trx(t, elements...)

	return signedTrx
}

type ContextFreeAction *pbantelope.Action

func Trx(t testing.T, elements ...interface{}) *pbantelope.Transaction {
	trx := &pbantelope.Transaction{}

	for _, element := range elements {
		switch v := element.(type) {
		case *pbantelope.Action:
			trx.Actions = append(trx.Actions, v)
		case ContextFreeAction:
			trx.ContextFreeActions = append(trx.ContextFreeActions, (*pbantelope.Action)(v))
		default:
			failInvalidComponent(t, "transaction", element)
		}
	}

	return trx
}

type ActionData string
type actionMatched struct {
	matched bool
	system  bool
}
type undecodedActionData bool

var ActionMatched = actionMatched{true, false}
var ActionSystemMatched = actionMatched{true, true}

var UndecodedActionData = undecodedActionData(true)

type ActionIndex uint32
type ExecutionIndex uint32
type GlobalSequence uint64

func ActionTrace(t testing.T, receiverAccountActionNameTriplet string, components ...interface{}) *pbantelope.ActionTrace {
	parts := strings.Split(receiverAccountActionNameTriplet, ":")
	receiver := parts[0]

	var account, actionName string
	if len(parts) == 2 {
		account = receiver
		actionName = parts[1]
	} else {
		// We assume 3 for now
		account = parts[1]
		actionName = parts[2]
	}

	actTrace := &pbantelope.ActionTrace{
		Receiver: receiver,
		Receipt: &pbantelope.ActionReceipt{
			Receiver: receiver,
		},
		Action: Action(t, account+":"+actionName, components...),
	}

	return transformActionTrace(t, actTrace, components)
}

func ActionTraceFail(t testing.T, tripletName string, components ...interface{}) *pbantelope.ActionTrace {
	components = append(components, GlobalSequence(0))
	out := ActionTrace(t, tripletName, components...)
	out.Receipt = nil

	return out
}

func ActionTraceSetABI(t testing.T, account string, abi *eos.ABI, components ...interface{}) *pbantelope.ActionTrace {
	var abiHex []byte
	var err error
	if abi != nil {
		abiHex, err = eos.MarshalBinary(abi)
		require.NoError(t, err)
	}

	setABI := &system.SetABI{Account: eos.AccountName(account), ABI: eos.HexBytes(abiHex)}
	rawData, err := eos.MarshalBinary(setABI)
	require.NoError(t, err)

	jsonData, err := json.Marshal(setABI)
	require.NoError(t, err)

	actTrace := &pbantelope.ActionTrace{
		Receiver: "eosio",
		Receipt: &pbantelope.ActionReceipt{
			Receiver: "eosio",
		},
		Action: &pbantelope.Action{
			Account:  "eosio",
			Name:     "setabi",
			JsonData: string(jsonData),
			RawData:  rawData,
		},
	}

	return transformActionTrace(t, actTrace, components)
}

func transformActionTrace(t testing.T, actTrace *pbantelope.ActionTrace, components []interface{}) *pbantelope.ActionTrace {
	ignoreIfActionComponent := ignoreComponent(func(component interface{}) bool {
		switch component.(type) {
		case actionComponent:
		case ActionData:
		default:
			return false
		}

		// Ignore all
		return true
	})

	for _, component := range components {
		switch v := component.(type) {
		case undecodedActionData:
			actTrace.Action.JsonData = ""
		case ExecutionIndex:
			actTrace.ExecutionIndex = uint32(v)
		case GlobalSequence:
			actTrace.Receipt.GlobalSequence = uint64(v)
		case actionMatched:
			actTrace.FilteringMatched = bool(v.matched)
			actTrace.FilteringMatchedSystemActionFilter = bool(v.system)
		default:
			failInvalidComponent(t, "action trace", component, ignoreIfActionComponent)
		}
	}

	return actTrace
}

func CFAAction(t testing.T, pairName string, abi *eos.ABI, data string) ContextFreeAction {
	return ContextFreeAction(Action(t, pairName, abi, data))
}

type Authorization string

func (a Authorization) apply(action *pbantelope.Action) {
	parts := strings.Split(string(a), "@")
	account := parts[0]

	permission := "active"
	if len(parts) > 1 {
		permission = parts[1]
	}

	action.Authorization = append(action.Authorization, &pbantelope.PermissionLevel{
		Actor:      account,
		Permission: permission,
	})
}

type actionComponent interface {
	apply(action *pbantelope.Action)
}

type authorizations []string

func Authorizations(elements ...string) authorizations {
	return authorizations(elements)
}

func (a authorizations) apply(action *pbantelope.Action) {
	for _, authorization := range a {
		Authorization(authorization).apply(action)
	}
}

func Action(t testing.T, pairName string, components ...interface{}) *pbantelope.Action {
	parts := strings.Split(pairName, ":")
	account := parts[0]
	actionName := parts[1]

	abi := findABIComponent(components)
	data := findActionData(components)

	var rawData []byte
	if abi != nil && data != "" {
		var err error
		rawData, err = abi.EncodeAction(eos.ActionName(actionName), []byte(data))
		require.NoError(t, err)
	}

	action := &pbantelope.Action{
		Account:  account,
		Name:     actionName,
		RawData:  rawData,
		JsonData: data,
	}

	for _, component := range components {
		switch v := component.(type) {
		case actionComponent:
			v.apply(action)
		case *eos.ABI:
		case ActionData:
		case actionMatched:
		case GlobalSequence:
			// Ignored
		default:
			failInvalidComponent(t, "action", component)
		}
	}

	return action
}

func findABIComponent(components []interface{}) *eos.ABI {
	if component := findComponent(components, func(component interface{}) bool { _, ok := component.(*eos.ABI); return ok }); component != nil {
		return component.(*eos.ABI)
	}

	return nil
}

func findActionData(components []interface{}) string {
	if component := findComponent(components, func(component interface{}) bool { _, ok := component.(ActionData); return ok }); component != nil {
		return string(component.(ActionData))
	}

	return ""
}

func findComponent(components []interface{}, doesMatch func(component interface{}) bool) interface{} {
	for _, component := range components {
		if doesMatch(component) {
			return component
		}
	}

	return nil
}

func findTypedComponent(components []interface{}, typeInfo interface{}) interface{} {
	searchedType := reflect.TypeOf(typeInfo)
	for _, component := range components {
		if reflect.TypeOf(component) == searchedType {
			return component
		}
	}

	return nil
}

func hasComponent(components []interface{}, doesMatch func(component interface{}) bool) bool {
	return findComponent(components, doesMatch) != nil
}

func TrxOp(t testing.T, signedTrx *pbantelope.SignedTransaction) *pbantelope.TrxOp {
	op := &pbantelope.TrxOp{
		Transaction: signedTrx,
	}

	return op
}

type DtrxOpPayer string

func DtrxOp(t testing.T, operation string, trxID string, components ...interface{}) *pbantelope.DTrxOp {
	opName := pbantelope.DTrxOp_Operation_value["OPERATION_"+strings.ToUpper(operation)]

	op := &pbantelope.DTrxOp{
		Operation:     pbantelope.DTrxOp_Operation(opName),
		TransactionId: trxID,
	}

	for _, component := range components {
		switch v := component.(type) {
		case ActionIndex:
			op.ActionIndex = uint32(v)
		case DtrxOpPayer:
			op.Payer = string(v)
		case *pbantelope.SignedTransaction:
			op.Transaction = v
		default:
			failInvalidComponent(t, "dtrx op", component)
		}
	}

	return op
}

func DBOp(t testing.T, op string, path string, payer string, data string, components ...interface{}) *pbantelope.DBOp {
	paths := strings.Split(path, "/")

	// Split those with → instead, will probably improve readability
	payers := strings.Split(payer, "/")
	datas := strings.Split(data, "/")

	op = strings.ToUpper(op)
	shortOpToLongOp := map[string]string{
		"INS": "INSERT",
		"UPD": "UPDATE",
		"REM": "REMOVE",
	}
	longOp, found := shortOpToLongOp[op]
	if found {
		op = longOp
	}

	dbOp := &pbantelope.DBOp{
		Operation:  pbantelope.DBOp_Operation(pbantelope.DBOp_Operation_value["OPERATION_"+op]),
		Code:       paths[0],
		TableName:  paths[1],
		Scope:      paths[2],
		PrimaryKey: paths[3],
	}

	if payers[0] != "" {
		dbOp.OldPayer = payers[0]
	}

	if payers[1] != "" {
		dbOp.NewPayer = payers[1]
	}

	var abi *eos.ABI
	for _, component := range components {
		switch v := component.(type) {
		case *eos.ABI:
			abi = v
		default:
			failInvalidComponent(t, "db op", component)
		}
	}

	dataToBinary := func(content string) []byte {
		if abi != nil {
			data, err := abi.EncodeTable(eos.TableName(dbOp.TableName), []byte(content))
			require.NoError(t, err)

			return data
		}

		return []byte(content)
	}

	if datas[0] != "" {
		dbOp.OldData = dataToBinary(datas[0])
	}

	if datas[1] != "" {
		dbOp.NewData = dataToBinary(datas[1])
	}

	return dbOp
}

type OldPerm *pbantelope.PermissionObject
type NewPerm *pbantelope.PermissionObject

func PermOp(t testing.T, op string, components ...interface{}) *pbantelope.PermOp {
	op = strings.ToUpper(op)
	shortOpToLongOp := map[string]string{
		"INS": "INSERT",
		"UPD": "UPDATE",
		"REM": "REMOVE",
	}
	longOp, found := shortOpToLongOp[op]
	if found {
		op = longOp
	}

	permOp := &pbantelope.PermOp{
		Operation: pbantelope.PermOp_Operation(pbantelope.PermOp_Operation_value["OPERATION_"+op]),
	}

	for _, component := range components {
		switch v := component.(type) {
		case OldPerm:
			permOp.OldPerm = v
		case NewPerm:
			permOp.OldPerm = v
		case ActionIndex:
			permOp.ActionIndex = uint32(v)
		default:
			failInvalidComponent(t, "perm op", component)
		}
	}

	return permOp
}

type PublicKey string

func Permission(t testing.T, accountPermission string, components ...interface{}) *pbantelope.PermissionObject {
	paths := strings.Split(accountPermission, "@")

	permission := &pbantelope.PermissionObject{
		Owner: paths[0],
		Name:  paths[1],
	}

	for _, component := range components {
		switch v := component.(type) {
		case PublicKey:
			keyWeight := &pbantelope.KeyWeight{PublicKey: string(v), Weight: 1}
			if permission.Authority == nil {
				permission.Authority = &pbantelope.Authority{}
			}

			permission.Authority.Keys = append(permission.Authority.Keys, keyWeight)
		default:
			failInvalidComponent(t, "permission object", component)
		}
	}

	return permission
}

func TableOp(t testing.T, op string, path string, payer string) *pbantelope.TableOp {
	paths := strings.Split(path, "/")

	return &pbantelope.TableOp{
		Operation: pbantelope.TableOp_Operation(pbantelope.TableOp_Operation_value["OPERATION_"+strings.ToUpper(op)]),
		Code:      paths[0],
		TableName: paths[1],
		Scope:     paths[2],
		Payer:     payer,
	}
}

type ignoreComponent func(v interface{}) bool

func failInvalidComponent(t testing.T, tag string, component interface{}, options ...interface{}) {
	shouldIgnore := ignoreComponent(func(v interface{}) bool { return false })
	for _, option := range options {
		switch v := option.(type) {
		case ignoreComponent:
			shouldIgnore = v
		}
	}

	if shouldIgnore(component) {
		return
	}

	require.FailNowf(t, "invalid component", "Invalid %s component of type %T", tag, component)
}

func logInvalidComponent(tag string, component interface{}) {
	zlog.Info(fmt.Sprintf("invalid %s component of type %T", tag, component))
}
