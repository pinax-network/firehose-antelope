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

package nodemanager

import (
	"regexp"
	"strings"

	logplugin "github.com/streamingfast/node-manager/log_plugin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logLevelRegex = regexp.MustCompile("^(<[0-9]>)?(debug|info|warn|error)")

func newToZapLogPlugin(debugDeepMind bool, logger *zap.Logger) *logplugin.ToZapLogPlugin {
	return logplugin.NewToZapLogPlugin(debugDeepMind, logger, logplugin.ToZapLogPluginLogLevel(newLogLevelExtractor().extract))
}

var discardRegex = regexp.MustCompile("(?i)" + "wabt.hpp:.*misaligned reference")
var toInfoRegex = regexp.MustCompile("(?i)" + "(" +
	strings.Join([]string{
		"controller.cpp:.*(No existing chain state or fork database|Initializing new blockchain with genesis state)",
		"platform_timer_accurac:.*Checktime timer",
		"net_plugin.cpp:.*closing connection to:",
		"net_plugin.cpp:.*connection failed to:",
		"CHAINBASE:*",
	}, "|") +
	")")

type logLevelExtractor struct {
	lastLineLevel zapcore.Level
}

func newLogLevelExtractor() *logLevelExtractor {
	return &logLevelExtractor{}
}

func (l *logLevelExtractor) extract(in string) zapcore.Level {

	if discardRegex.MatchString(in) {
		l.lastLineLevel = logplugin.NoDisplay
		return l.lastLineLevel
	}

	if toInfoRegex.MatchString(in) {
		l.lastLineLevel = zap.InfoLevel
		return l.lastLineLevel
	}

	groups := logLevelRegex.FindStringSubmatch(in)
	if len(groups) <= 2 {
		// nodeos has multi line logs where only the first line contains the log level. This is likely the case here,
		// so we return the log level of the last line instead to avoid losing log output.
		return l.lastLineLevel
	}

	switch groups[2] {
	case "debug":
		l.lastLineLevel = zap.DebugLevel
	case "info":
		l.lastLineLevel = zap.InfoLevel
	case "warn":
		l.lastLineLevel = zap.WarnLevel
	case "error":
		l.lastLineLevel = zap.ErrorLevel
	default:
		l.lastLineLevel = zap.DebugLevel
	}

	return l.lastLineLevel
}
