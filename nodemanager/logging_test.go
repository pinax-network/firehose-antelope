package nodemanager

import (
	"testing"

	"github.com/streamingfast/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToZapLogPlugin_LogLevel(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		out  []string
	}{
		// The standard `nodeos` output
		{
			"debug",
			[]string{"debug  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."},
			[]string{`{"level":"debug","msg":"debug  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."}`},
		},
		{
			"info",
			[]string{"info  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."},
			[]string{`{"level":"info","msg":"info  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."}`},
		},
		{
			"warn",
			[]string{"warn  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."},
			[]string{`{"level":"warn","msg":"warn  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."}`},
		},
		{
			"error",
			[]string{"error  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."},
			[]string{`{"level":"error","msg":"error  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."}`},
		},
		// this is now returning the previous log level, due to handling of multi line logs
		//{
		//	"other",
		//	"other  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ...",
		//	`{"level":"debug","msg":"other  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."}`,
		//},

		// multi line nodeos output
		{
			"warn multi line",
			[]string{
				"warn  2022-12-07T13:01:16.666 nodeos    chain_plugin.cpp:1246         plugin_initialize ...",
				"Cannot load snapshot, latest does not exist",
				`    {"name":"latest"}`,
				"    nodeos  chain_plugin.cpp:902 plugin_initialize"},
			[]string{
				`{"level":"warn","msg":"warn  2022-12-07T13:01:16.666 nodeos    chain_plugin.cpp:1246         plugin_initialize ..."}`,
				`{"level":"warn","msg":"Cannot load snapshot, latest does not exist"}`,
				`{"level":"warn","msg":"    {\"name\":\"latest\"}"}`,
				`{"level":"warn","msg":"    nodeos  chain_plugin.cpp:902 plugin_initialize"}`},
		},

		// The weird `nodeos` output where some markup is appended
		{
			"weird markup, debug",
			[]string{"<0>debug  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."},
			[]string{`{"level":"debug","msg":"<0>debug  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."}`},
		},
		{
			"weird markup, info",
			[]string{"<6>info  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."},
			[]string{`{"level":"info","msg":"<6>info  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."}`},
		},
		{
			"weird markup, warn",
			[]string{"<4>warn  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."},
			[]string{`{"level":"warn","msg":"<4>warn  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."}`},
		},
		{
			"weird markup, error",
			[]string{"<3>error  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."},
			[]string{`{"level":"error","msg":"<3>error  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."}`},
		},
		// this is now returning the previous log level, due to handling of multi line logs
		//{
		//	"weird markup, other",
		//	"[random]other  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ...",
		//	`{"level":"debug","msg":"[random]other  2020-10-05T13:53:11.749 thread-0  http_plugin.cpp:895 ..."}`,
		//},

		{
			"discarded wabt reference",
			[]string{"warn  2020-10-05T13:53:11.749 thread-5  wabt.hpp:106	misaligned reference"},
			[]string{},
		},
		{
			"to info closing connection to",
			[]string{"error  2020-10-05T13:53:11.749 thread-0  net_plugin.cpp:253 Closing connection to: 192.168.0.189"},
			[]string{`{"level":"info","msg":"error  2020-10-05T13:53:11.749 thread-0  net_plugin.cpp:253 Closing connection to: 192.168.0.189"}`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wrapper := logging.NewTestLogger(t)
			plugin := newToZapLogPlugin(false, wrapper.Instance())

			for _, line := range test.in {
				plugin.LogLine(line)
			}

			loggedLines := wrapper.RecordedLines(t)

			if len(test.out) == 0 {
				require.Len(t, loggedLines, 0)
			} else {
				require.Len(t, loggedLines, len(test.out))

				for i, out := range test.out {
					assert.Equal(t, out, loggedLines[i])
				}
			}
		})
	}
}
