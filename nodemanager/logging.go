package nodemanager

import (
	"regexp"
	"strings"

	logplugin "github.com/streamingfast/node-manager/log_plugin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logLevelRegex = regexp.MustCompile("^(INFO|WARN|ERROR)")

func newToZapLogPlugin(debugDeepMind bool, logger *zap.Logger) *logplugin.ToZapLogPlugin {
	return logplugin.NewToZapLogPlugin(debugDeepMind, logger, logplugin.ToZapLogPluginLogLevel(logLevelExtractor))
}

func logLevelExtractor(in string) zapcore.Level {
	if strings.Contains(in, "Upgrade blockchain database version") {
		return zap.InfoLevel
	}

	groups := logLevelRegex.FindStringSubmatch(in)
	if len(groups) <= 1 {
		return zap.DebugLevel
	}

	switch groups[1] {
	case "INFO":
		return zap.InfoLevel
	case "WARN":
		return zap.WarnLevel
	case "ERROR":
		return zap.ErrorLevel
	default:
		return zap.DebugLevel
	}
}
