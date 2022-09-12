package codec

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/firehose-acme/types"
	pbacme "github.com/streamingfast/firehose-acme/types/pb/sf/acme/type/v1"
	"go.uber.org/zap"
)

// ConsoleReader is what reads the `geth` output directly. It builds
// up some LogEntry objects. See `LogReader to read those entries .
type ConsoleReader struct {
	lines chan string
	close func()

	ctx  *parseCtx
	done chan interface{}

	logger *zap.Logger
}

func NewConsoleReader(logger *zap.Logger, lines chan string) (*ConsoleReader, error) {
	l := &ConsoleReader{
		lines:  lines,
		close:  func() {},
		done:   make(chan interface{}),
		logger: logger,
	}
	return l, nil
}

//todo: WTF?
func (r *ConsoleReader) Done() <-chan interface{} {
	return r.done
}

func (r *ConsoleReader) Close() {
	r.close()
}

type parsingStats struct {
	startAt  time.Time
	blockNum uint64
	data     map[string]int
	logger   *zap.Logger
}

func newParsingStats(logger *zap.Logger, block uint64) *parsingStats {
	return &parsingStats{
		startAt:  time.Now(),
		blockNum: block,
		data:     map[string]int{},
		logger:   logger,
	}
}

func (s *parsingStats) log() {
	s.logger.Info("reader block stats",
		zap.Uint64("block_num", s.blockNum),
		zap.Int64("duration", int64(time.Since(s.startAt))),
		zap.Reflect("stats", s.data),
	)
}

func (s *parsingStats) inc(key string) {
	if s == nil {
		return
	}
	k := strings.ToLower(key)
	value := s.data[k]
	value++
	s.data[k] = value
}

type parseCtx struct {
	currentBlock *pbacme.Block
	stats        *parsingStats

	logger *zap.Logger
}

func newContext(logger *zap.Logger, height uint64) *parseCtx {
	return &parseCtx{
		currentBlock: &pbacme.Block{
			Height:       height,
			Transactions: []*pbacme.Transaction{},
		},
		stats: newParsingStats(logger, height),

		logger: logger,
	}
}

func (r *ConsoleReader) ReadBlock() (out *bstream.Block, err error) {
	block, err := r.next()
	if err != nil {
		return nil, err
	}

	return types.BlockFromProto(block)
}

const (
	LogPrefix     = "FIRE"
	LogBeginBlock = "BLOCK_BEGIN"
	LogBeginTrx   = "BEGIN_TRX"
	LogBeginEvent = "TRX_BEGIN_EVENT"
	LogEventAttr  = "TRX_EVENT_ATTR"
	LogEndTrx     = "END_TRX"
	LogEndBlock   = "BLOCK_END"
)

func (r *ConsoleReader) next() (out *pbacme.Block, err error) {
	for line := range r.lines {
		if !strings.HasPrefix(line, LogPrefix) {
			continue
		}

		// This code assumes that distinct element do not contains space. This can happen
		// for example when exchanging JSON object (although we strongly discourage usage of
		// JSON, use serialized Protobuf object). If you happen to have spaces in the last element,
		// refactor the code here to avoid the split and perform the split in the line handler directly
		// instead.
		tokens := strings.Split(line[len(LogPrefix)+1:], " ")
		if len(tokens) < 2 {
			return nil, fmt.Errorf("invalid log line %q, expecting at least two tokens", line)
		}

		// Order the case from most occurring line prefix to least occurring
		switch tokens[0] {
		case LogBeginEvent:
			err = r.ctx.eventBegin(tokens[1:])
		case LogEventAttr:
			err = r.ctx.eventAttr(tokens[1:])
		case LogBeginTrx:
			err = r.ctx.trxBegin(tokens[1:])
		case LogBeginBlock:
			err = r.blockBegin(tokens[1:])
		case LogEndBlock:
			// This end the execution of the reading loop as we have a full block here
			return r.ctx.readBlockEnd(tokens[1:])
		default:
			if r.logger.Core().Enabled(zap.DebugLevel) {
				r.logger.Debug("skipping unknown deep mind log line", zap.String("line", line))
			}
			continue
		}

		if err != nil {
			chunks := strings.SplitN(line, " ", 2)
			return nil, fmt.Errorf("%s: %w (line %q)", chunks[0], err, line)
		}
	}

	r.logger.Info("lines channel has been closed")
	return nil, io.EOF
}

func (r *ConsoleReader) processData(reader io.Reader) error {
	scanner := r.buildScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		r.lines <- line
	}

	if scanner.Err() == nil {
		close(r.lines)
		return io.EOF
	}

	return scanner.Err()
}

func (r *ConsoleReader) buildScanner(reader io.Reader) *bufio.Scanner {
	buf := make([]byte, 50*1024*1024)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(buf, 50*1024*1024)

	return scanner
}

// Format:
// FIRE BLOCK_BEGIN <NUM>
func (r *ConsoleReader) blockBegin(params []string) error {
	if err := validateChunk(params, 1); err != nil {
		return fmt.Errorf("invalid log line length: %w", err)
	}

	blockHeight, err := strconv.ParseUint(params[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid block num: %w", err)
	}

	//Push new block meta
	r.ctx = newContext(r.logger, blockHeight)
	return nil
}

// Format:
// FIRE BLOCK_BEGIN <HASH> <TYPE> <FROM> <TO> <AMOUNT> <FEE> <SUCCESS>
func (ctx *parseCtx) trxBegin(params []string) error {
	if err := validateChunk(params, 7); err != nil {
		return fmt.Errorf("invalid log line length: %w", err)
	}
	if ctx == nil {
		return fmt.Errorf("did not process a BLOCK_BEGIN")
	}

	trx := &pbacme.Transaction{
		Type:     params[1],
		Hash:     params[0],
		Sender:   params[2],
		Receiver: params[3],
		Success:  params[6] == "true",
		Events:   []*pbacme.Event{},
	}

	v, ok := new(big.Int).SetString(params[4], 16)
	if !ok {
		return fmt.Errorf("unable to parse trx amount %s", params[4])
	}
	trx.Amount = &pbacme.BigInt{Bytes: v.Bytes()}

	v, ok = new(big.Int).SetString(params[5], 16)
	if !ok {
		return fmt.Errorf("unable to parse trx amount %s", params[4])
	}
	trx.Fee = &pbacme.BigInt{Bytes: v.Bytes()}

	ctx.currentBlock.Transactions = append(ctx.currentBlock.Transactions, trx)
	return nil
}

// Format:
// FIRE TRX_BEGIN_EVENT <TRX_HASH> <TYPE>

func (ctx *parseCtx) eventBegin(params []string) error {
	if err := validateChunk(params, 2); err != nil {
		return fmt.Errorf("invalid log line length: %w", err)
	}
	if ctx == nil {
		return fmt.Errorf("did not process a BLOCK_BEGIN")
	}
	if len(ctx.currentBlock.Transactions) == 0 {
		return fmt.Errorf("did not process a BEGIN_TRX")
	}

	trx := ctx.currentBlock.Transactions[len(ctx.currentBlock.Transactions)-1]
	if trx.Hash != params[0] {
		return fmt.Errorf("last transaction hash %q does not match the event trx hash %q", trx.Hash, params[0])
	}

	trx.Events = append(trx.Events, &pbacme.Event{
		Type:       params[1],
		Attributes: []*pbacme.Attribute{},
	})

	ctx.currentBlock.Transactions[len(ctx.currentBlock.Transactions)-1] = trx
	return nil
}

// Format:
// FIRE TRX_EVENT_ATTR <TRX_HASH> <EVENT_INDEX> <KEY> <VALUE>
func (ctx *parseCtx) eventAttr(params []string) error {
	if err := validateChunk(params, 4); err != nil {
		return fmt.Errorf("invalid log line length: %w", err)
	}
	if ctx == nil {
		return fmt.Errorf("did not process a BLOCK_BEGIN")
	}
	if len(ctx.currentBlock.Transactions) == 0 {
		return fmt.Errorf("did not process a BEGIN_TRX")
	}

	trx := ctx.currentBlock.Transactions[len(ctx.currentBlock.Transactions)-1]
	if trx.Hash != params[0] {
		return fmt.Errorf("last transaction hash %q does not match the event trx hash %q", trx.Hash, params[0])
	}

	eventIndex, err := strconv.ParseUint(params[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid event index: %w", err)
	}

	if len(trx.Events) < int(eventIndex) {
		return fmt.Errorf("length of events array does not match event index: %d", eventIndex)
	}
	event := trx.Events[eventIndex]
	event.Attributes = append(event.Attributes, &pbacme.Attribute{
		Key:   params[2],
		Value: params[3],
	})
	trx.Events[eventIndex] = event
	ctx.currentBlock.Transactions[len(ctx.currentBlock.Transactions)-1] = trx
	return nil
}

// Format:
// FIRE BLOCK_END <HEIGHT> <HASH> <PREV_HASH> <TIMESTAMP> <TRX-COUNT>
func (ctx *parseCtx) readBlockEnd(params []string) (*pbacme.Block, error) {
	if err := validateChunk(params, 5); err != nil {
		return nil, fmt.Errorf("invalid log line length: %w", err)
	}

	if ctx.currentBlock == nil {
		return nil, fmt.Errorf("current block not set")
	}

	blockHeight, err := strconv.ParseUint(params[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse blockNum: %w", err)
	}
	if blockHeight != ctx.currentBlock.Height {
		return nil, fmt.Errorf("end block height does not match active block height, got block height %d but current is block height %d", blockHeight, ctx.currentBlock.Height)
	}

	trxCount, err := strconv.ParseUint(params[4], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse blockNum: %w", err)
	}

	if len(ctx.currentBlock.Transactions) != int(trxCount) {
		return nil, fmt.Errorf("expected %d transaction count, got %d", trxCount, len(ctx.currentBlock.Transactions))
	}

	timestamp, err := strconv.ParseUint(params[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block timestamp: %w", err)
	}

	ctx.currentBlock.Hash = params[1]
	ctx.currentBlock.PrevHash = params[2]
	ctx.currentBlock.Timestamp = timestamp

	ctx.logger.Debug("console reader read block",
		zap.Uint64("height", ctx.currentBlock.Height),
		zap.String("hash", ctx.currentBlock.Hash),
		zap.String("prev_hash", ctx.currentBlock.PrevHash),
		zap.Int("trx_count", len(ctx.currentBlock.Transactions)),
	)

	return ctx.currentBlock, nil
}

func validateChunk(params []string, count int) error {
	if len(params) != count {
		return fmt.Errorf("%d fields required but found %d", count, len(params))
	}
	return nil
}
