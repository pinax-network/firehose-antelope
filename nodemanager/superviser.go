package nodemanager

import (
	"strconv"
	"strings"
	"sync"

	"github.com/ShinyTrinkets/overseer"
	nodeManager "github.com/streamingfast/node-manager"
	logplugin "github.com/streamingfast/node-manager/log_plugin"
	"github.com/streamingfast/node-manager/superviser"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Superviser struct {
	*superviser.Superviser

	infoMutex           sync.Mutex
	binary              string
	arguments           []string
	dataDir             string
	lastBlockSeen       uint64
	serverId            string
	headBlockUpdateFunc nodeManager.HeadBlockUpdater
	Logger              *zap.Logger
}

func (s *Superviser) GetName() string {
	return "acme"
}

func NewSuperviser(
	binary string,
	arguments []string,
	dataDir string,
	headBlockUpdateFunc nodeManager.HeadBlockUpdater,
	debugFirehose bool,
	logToZap bool,
	appLogger *zap.Logger,
	nodelogger *zap.Logger,
) *Superviser {
	// Ensure process manager line buffer is large enough (50 MiB) for our Deep Mind instrumentation outputting lot's of text.
	overseer.DEFAULT_LINE_BUFFER_SIZE = 50 * 1024 * 1024

	supervisor := &Superviser{
		Superviser:          superviser.New(appLogger, binary, arguments),
		Logger:              appLogger,
		binary:              binary,
		arguments:           arguments,
		dataDir:             dataDir,
		headBlockUpdateFunc: headBlockUpdateFunc,
	}

	supervisor.RegisterLogPlugin(logplugin.LogPluginFunc(supervisor.lastBlockSeenLogPlugin))

	if logToZap {
		supervisor.RegisterLogPlugin(newToZapLogPlugin(debugFirehose, nodelogger))
	} else {
		supervisor.RegisterLogPlugin(logplugin.NewToConsoleLogPlugin(debugFirehose))
	}

	appLogger.Info("created acme superviser", zap.Object("superviser", supervisor))
	return supervisor
}

func (s *Superviser) GetCommand() string {
	return s.binary + " " + strings.Join(s.arguments, " ")
}

func (s *Superviser) LastSeenBlockNum() uint64 {
	return s.lastBlockSeen
}

func (s *Superviser) ServerID() (string, error) {
	return s.serverId, nil
}

func (s *Superviser) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("binary", s.binary)
	enc.AddArray("arguments", stringArray(s.arguments))
	enc.AddString("data_dir", s.dataDir)
	enc.AddUint64("last_block_seen", s.lastBlockSeen)
	enc.AddString("server_id", s.serverId)

	return nil
}

func (s *Superviser) lastBlockSeenLogPlugin(line string) {
	if !strings.HasPrefix(line, "FIRE BLOCK_BEGIN") {
		return
	}

	blockNumStr := line[18:]

	blockNum, err := strconv.ParseUint(blockNumStr, 10, 64)
	if err != nil {
		s.Logger.Error("unable to extract last block num",
			zap.String("line", line),
			zap.String("block_num_str", blockNumStr),
			zap.Error(err),
		)
		return
	}

	s.lastBlockSeen = blockNum
}
