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

var logLevelRegex = regexp.MustCompile("^(<[0-9]>)?(info|warn|error)")

func newToZapLogPlugin(debugDeepMind bool, logger *zap.Logger) *logplugin.ToZapLogPlugin {
	return logplugin.NewToZapLogPlugin(debugDeepMind, logger, logplugin.ToZapLogPluginLogLevel(logLevelExtractor))
}

var discardRegex = regexp.MustCompile("(?i)" + "wabt.hpp:.*misaligned reference")
var toInfoRegex = regexp.MustCompile("(?i)" + "(" +
	strings.Join([]string{
		"controller.cpp:.*(No existing chain state or fork database|Initializing new blockchain with genesis state)",
		"platform_timer_accurac:.*Checktime timer",
		"net_plugin.cpp:.*closing connection to:",
		"net_plugin.cpp:.*connection failed to:",
	}, "|") +
	")")

func logLevelExtractor(in string) zapcore.Level {
	if discardRegex.MatchString(in) {
		return logplugin.NoDisplay
	}

	if toInfoRegex.MatchString(in) {
		return zap.InfoLevel
	}

	groups := logLevelRegex.FindStringSubmatch(in)
	if len(groups) <= 2 {
		return zap.DebugLevel
	}

	switch groups[2] {
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.DebugLevel
	}
}
