package cli

import (
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

var userLog = launcher.UserLog
var zlog *zap.Logger

func init() {
	logging.Register("github.com/streamingfast/dfuse-ethereum/cmd/dfuseeth", &zlog)
}
