package cli

import (
	"fmt"
	"strings"
	"time"

	// Needs to be in this file which is the main entry of wrapper binary
	_ "github.com/streamingfast/dauth/authenticator/null"   // auth null plugin
	_ "github.com/streamingfast/dauth/authenticator/secret" // auth secret/hard-coded plugin
	_ "github.com/streamingfast/dauth/ratelimiter/null"     // ratelimiter plugin

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dlauncher/flags"
	"github.com/streamingfast/dlauncher/launcher"
	"go.uber.org/zap"
)

var RootCmd = &cobra.Command{Use: "fireacme", Short: "Acme on StreamingFast"}
var allFlags = make(map[string]bool) // used as global because of async access to cobra init functions

func Main() {
	cobra.OnInitialize(func() {
		allFlags = flags.AutoBind(RootCmd, "fireacme")
	})

	RootCmd.PersistentFlags().StringP("data-dir", "d", "./sf-data", "Path to data storage for all components of dfuse")
	RootCmd.PersistentFlags().StringP("config-file", "c", "./sf.yaml", "Configuration file to use. No config file loaded if set to an empty string.")
	RootCmd.PersistentFlags().String("log-format", "text", "Format for logging to stdout. Either 'text' or 'stackdriver'")
	RootCmd.PersistentFlags().Bool("log-to-file", true, "Also write logs to {data-dir}/dfuse.log.json ")
	RootCmd.PersistentFlags().CountP("verbose", "v", "Enables verbose output (-vvvv for max verbosity)")

	RootCmd.PersistentFlags().String("log-level-switcher-listen-addr", "localhost:1065", "If non-empty, the process will listen on this address for json-formatted requests to change different logger levels (see DEBUG.md for more info)")
	RootCmd.PersistentFlags().String("metrics-listen-addr", MetricsListenAddr, "If non-empty, the process will listen on this address to server Prometheus metrics")
	RootCmd.PersistentFlags().String("pprof-listen-addr", "localhost:6060", "If non-empty, the process will listen on this address for pprof analysis (see https://golang.org/pkg/net/http/pprof/)")
	RootCmd.PersistentFlags().Duration("startup-delay", 0, "delay before launching dfuse process")

	derr.Check("registering application flags", launcher.RegisterFlags(StartCmd))

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
			zlog.Info("sleeping before starting apps", zap.Duration("delay", startupDelay))
			time.Sleep(startupDelay)
		}
		return nil
	}

	derr.Check("acme-blockchain", RootCmd.Execute())
}

var startCmdExample = `fireacme start mindreader`
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
