package nodemanager

import (
	"regexp"

	logplugin "github.com/streamingfast/node-manager/log_plugin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// This file configures a logging reader that transforms log lines received from the blockchain process running
// and then logs them inside the Firehose stack logging system.
//
// A simple regex is going to identify the level of the line and turn it into our internal level value.
//
// You **must** adapt this line to fit with the log lines of your chain. For example, the dummy blockchain we
// instrumented in `firehose-acme`, log lines look like:
//
//    time="2022-03-04T12:49:34-05:00" level=info msg="initializing node"
//
// So our regex look like the one below, extracting the `info` value from a group in the regexp.
var logLevelRegex = regexp.MustCompile("level=(debug|info|warn|warning|error)")

func newToZapLogPlugin(debugFirehose bool, logger *zap.Logger) *logplugin.ToZapLogPlugin {
	return logplugin.NewToZapLogPlugin(debugFirehose, logger, logplugin.ToZapLogPluginLogLevel(logLevelReader), logplugin.ToZapLogPluginTransformer(stripTimeTransformer))
}

func logLevelReader(in string) zapcore.Level {
	// If the regex does not match the line, log to `INFO` so at least we see something by default.
	groups := logLevelRegex.FindStringSubmatch(in)
	if len(groups) <= 1 {
		return zap.InfoLevel
	}

	switch groups[1] {
	case "debug", "DEBUG":
		return zap.DebugLevel
	case "info", "INFO":
		return zap.InfoLevel
	case "warn", "warning", "WARN", "WARNING":
		return zap.WarnLevel
	case "error", "ERROR":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

var timeRegex = regexp.MustCompile(`time="[0-9]{4}-[^"]+"\s*`)

func stripTimeTransformer(in string) string {
	return timeRegex.ReplaceAllString(in, "")
}
