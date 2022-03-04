package codec

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	pbcodec "github.com/streamingfast/firehose-acme/pb/sf/acme/codec/v1"
	"go.uber.org/zap"
)

// ConsoleReader is what reads the `geth` output directly. It builds
// up some LogEntry objects. See `LogReader to read those entries .
type ConsoleReader struct {
	lines chan string
	close func()

	ctx  *parseCtx
	done chan interface{}
}

func NewConsoleReader(lines chan string, rpcUrl string) (*ConsoleReader, error) {
	l := &ConsoleReader{
		lines: lines,
		close: func() {},
		done:  make(chan interface{}),
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
}

func newParsingStats(block uint64) *parsingStats {
	return &parsingStats{
		startAt:  time.Now(),
		blockNum: block,
		data:     map[string]int{},
	}
}

func (s *parsingStats) log() {
	zlog.Info("mindreader block stats",
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
	Height       uint64
	currentBlock *pbcodec.Block

	stats *parsingStats
}

func newContext(height uint64) *parseCtx {
	return &parseCtx{
		Height: height,
	}

}

func (r *ConsoleReader) Read() (out interface{}, err error) {
	return r.next()
}

const (
	LogPrefix     = "DMLOG"
	LogBeginBlock = "BLOCK_BEGIN"
	LogBlock      = "BLOCK_DATA"
	LogEndBlock   = "BLOCK_END"
)

func (r *ConsoleReader) next() (out interface{}, err error) {
	for line := range r.lines {
		if !strings.HasPrefix(line, LogPrefix) {
			continue
		}

		tokens := strings.Split(line[len(LogPrefix)+1:], " ")
		if len(tokens) < 2 {
			return nil, fmt.Errorf("invalid log line format: %s", line)
		}

		switch tokens[0] {
		case LogBeginBlock:
			err = r.blockBegin(tokens[1:])
		case LogBlock:
			err = r.ctx.readBlock(tokens[1:])
		case LogEndBlock:
			return r.ctx.readBlockEnd(tokens[1:])
		default:
			if tracer.Enabled() {
				zlog.Debug("skipping unknown deep mind log line", zap.String("line", line))
			}
			continue
		}

		if err != nil {
			chunks := strings.SplitN(line, " ", 2)
			return nil, fmt.Errorf("%s: %s (line %q)", chunks[0], err, line)
		}
	}

	zlog.Info("lines channel has been closed")
	return nil, io.EOF
}

func (r *ConsoleReader) ProcessData(reader io.Reader) error {
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
// DMLOG BLOCK_BEGIN <NUM>
func (r *ConsoleReader) blockBegin(params []string) error {
	if err := validateChunk(params, 1); err != nil {
		return fmt.Errorf("invalid log line length: %w", err)
	}

	blockHeight, err := strconv.ParseUint(params[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid block num: %w", err)
	}

	//Push new block meta
	r.ctx = newContext(blockHeight)
	return nil
}

// Format:
// DMLOG BLOCK <PROTOBUf-BLOCK>
func (ctx *parseCtx) readBlock(params []string) error {
	if err := validateChunk(params, 1); err != nil {
		return fmt.Errorf("invalid log line length: %w", err)
	}

	if ctx == nil {
		return fmt.Errorf("did not process a BLOCK_BEGIN")
	}

	block := &pbcodec.Block{}
	buf, err := base64.StdEncoding.DecodeString(params[0])
	if err != nil {
		return fmt.Errorf("unable to base64 decode block: %w", err)
	}
	if err := proto.Unmarshal(buf, block); err != nil {
		return fmt.Errorf("unable unmarshall block: %w", err)
	}

	ctx.currentBlock = block
	return nil
}

func (ctx *parseCtx) readBlockEnd(params []string) (*pbcodec.Block, error) {
	if err := validateChunk(params, 1); err != nil {
		return nil, fmt.Errorf("invalid log line length: %w", err)
	}

	if ctx.currentBlock == nil {
		return nil, fmt.Errorf("current block not set")
	}

	blockHeight, err := strconv.ParseUint(params[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse blockNum: %s", err)
	}
	if blockHeight != ctx.Height {
		return nil, fmt.Errorf("end block height does not match active block height, got block height %d but current is block height %d", blockHeight, ctx.Height)
	}

	zlog.Debug("console reader read block",
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
