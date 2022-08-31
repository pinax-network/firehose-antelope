package cli

import (
	"fmt"
	"strings"
	"time"

	// Needs to be in this file which is the main entry of wrapper binary
	_ "github.com/streamingfast/dauth/authenticator/null"   // auth null plugin
	_ "github.com/streamingfast/dauth/authenticator/secret" // auth secret/hard-coded plugin
	_ "github.com/streamingfast/dauth/ratelimiter/null"     // ratelimiter plugin
	"github.com/streamingfast/logging"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dlauncher/flags"
	"github.com/streamingfast/dlauncher/launcher"
	"go.uber.org/zap"
)

var rootLog, _ = logging.RootLogger("fireacme", "github.com/streamingfast/firehose-acme/cmd/fireacme/cli")

var RootCmd = &cobra.Command{Use: "fireacme", Short: "Acme on StreamingFast"}
var allFlags = make(map[string]bool) // used as global because of async access to cobra init functions

func Main() {
	cobra.OnInitialize(func() {
		allFlags = flags.AutoBind(RootCmd, "fireacme")
	})

	RootCmd.PersistentFlags().StringP("data-dir", "d", "./firehose-data", "Path to data storage for all components of the Firehose stack")
	RootCmd.PersistentFlags().StringP("config-file", "c", "./firehose.yaml", "Configuration file to use. No config file loaded if set to an empty string.")

	RootCmd.PersistentFlags().String("log-format", "text", "Format for logging to stdout. Either 'text' or 'stackdriver'")
	RootCmd.PersistentFlags().Bool("log-to-file", true, "Also write logs to {data-dir}/firehose.log.json ")
	RootCmd.PersistentFlags().String("log-level-switcher-listen-addr", "localhost:1065", FlagDescription(`
		If non-empty, a JSON based HTTP server will listen on this address to let you switch the default logging level
		of all registered loggers to a different one on the fly. This enables switching to debug level on
		a live running production instance. Use 'curl -XPUT -d '{"level":"debug","inputs":"*"} http://localhost:1065' to
		switch the level for all loggers. Each logger (even in transitive dependencies, at least those part of the core
		StreamingFast's Firehose) are registered using two identifiers, the overarching component usually all loggers in a
		library uses the same component name like 'bstream' or 'merger', and a fully qualified ID which is usually the Go
		package fully qualified name in which the logger is defined. The 'inputs' can be either one or many component's name
		like 'bstream|merger|firehose' or a regex that is matched against the fully qualified name. If there is a match for a
		given logger, it will change its level to the one specified in 'level' field. The valid levels are 'trace', 'debug',
		'info', 'warn', 'error', 'panic'. Can be used to silence loggers by using 'panic' (well, technically it's not a full
		silence but almost), or make them more verbose and change it back later.
	`))
	RootCmd.PersistentFlags().CountP("log-verbosity", "v", "Enables verbose output (-vvvv for max verbosity)")

	RootCmd.PersistentFlags().String("metrics-listen-addr", MetricsListenAddr, "If non-empty, the process will listen on this address to server the Prometheus metrics collected by the components.")
	RootCmd.PersistentFlags().String("pprof-listen-addr", "localhost:6060", "If non-empty, the process will listen on this address for pprof analysis (see https://golang.org/pkg/net/http/pprof/)")
	RootCmd.PersistentFlags().Duration("startup-delay", 0, FlagDescription(`
		Delay before launching the components defined in config file or via the command line arguments. This can be used to perform
		maintenance operations on a running container or pod prior it will actually start processing. Useful for example to clear
		a persistent disks of its content before starting, cleary cached content to try to resolve bugs, etc.
	`))

	derr.Check("registering application flags", launcher.RegisterFlags(rootLog, StartCmd))

	var availableCmds []string
	for app := range launcher.AppRegistry {
		availableCmds = append(availableCmds, app)
	}
	StartCmd.SetHelpTemplate(fmt.Sprintf(startCmdHelpTemplate, strings.Join(availableCmds, "\n  ")))
	StartCmd.Example = startCmdExample

	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := setupCmd(cmd); err != nil {
			return err
		}
		startupDelay := viper.GetDuration("global-startup-delay")
		if startupDelay.Microseconds() > 0 {
			rootLog.Info("sleeping before starting apps", zap.Duration("delay", startupDelay))
			time.Sleep(startupDelay)
		}
		return nil
	}

	derr.Check("acme-blockchain", RootCmd.Execute())
}

var startCmdExample = `fireacme start reader-node`
var startCmdHelpTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}} [all|command1 [command2...]]{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
  {{.Example}}{{end}}

Available Commands:
  %s{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
