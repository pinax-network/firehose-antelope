package cli

import (
	"fmt"
	"github.com/pinax-network/firehose-antelope/codec"
	"github.com/pinax-network/firehose-antelope/nodemanager"
	"github.com/streamingfast/bstream/blockstream"
	nodeManagerApp "github.com/streamingfast/node-manager/app/node_manager2"
	reader "github.com/streamingfast/node-manager/mindreader"
	"github.com/streamingfast/node-manager/operator"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	pbheadinfo "github.com/streamingfast/pbgo/sf/headinfo/v1"
	"google.golang.org/grpc"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/logging"
	nodeManager "github.com/streamingfast/node-manager"
	"github.com/streamingfast/node-manager/metrics"
	"go.uber.org/zap"
)

var readerLogger, readerTracer = logging.PackageLogger("reader", "github.com/pinax-network/firehose-antelope/reader")
var readerAcmeChainLogger, _ = logging.PackageLogger("reader.nodeos", "github.com/pinax-network/firehose-antelope/reader/antelope", DefaultLevelInfo)

func registerCommonNodeFlags(cmd *cobra.Command, flagPrefix string, managerAPIAddr string) {
	cmd.Flags().String(flagPrefix+"path", ChainExecutableName, FlagDescription(`
		Process that will be invoked to sync the chain, can be a full path or just the binary's name, in which case the binary is
		searched for paths listed by the PATH environment variable (following operating system rules around PATH handling).
	`))
	cmd.Flags().String(flagPrefix+"data-dir", "{sf-data-dir}/{node-role}/data", "nodeos data directory containing the blocks.log and state files")
	cmd.Flags().String(flagPrefix+"config-dir", "{sf-data-dir}/{node-role}/config", "nodeos config directory containing the config.ini and genesis.json files")
	cmd.Flags().Bool(flagPrefix+"debug-firehose-logs", false, "[DEV] Prints firehose instrumentation logs to standard output, should be use for debugging purposes only")
	cmd.Flags().Bool(flagPrefix+"log-to-zap", true, FlagDescription(`
		When sets to 'true', all standard error output emitted by the invoked process defined via '%s'
		is intercepted, split line by line and each line is then transformed and logged through the Firehose stack
		logging system. The transformation extracts the level and remove the timestamps creating a 'sanitized' version
		of the logs emitted by the blockchain's managed client process. If this is not desirable, disabled the flag
		and all the invoked process standard error will be redirect to 'fireantelope' standard's output.
	`, flagPrefix+"path"))
	cmd.Flags().String(flagPrefix+"manager-api-addr", managerAPIAddr, "Acme node manager API address")
	cmd.Flags().Duration(flagPrefix+"readiness-max-latency", 30*time.Second, "Determine the maximum head block latency at which the instance will be determined healthy. Some chains have more regular block production than others.")
	cmd.Flags().String(flagPrefix+"arguments", "", "Extra arguments to be passed when executing nodeos binary.")
	cmd.Flags().String(flagPrefix+"bootstrap-snapshot-url", "", FlagDescription(`
		Snapshot file URL to bootstrap nodeos from. If this is set the snapshot will be downloaded to the data directory and 
		nodeos will be started with the --snapshot flag. This can be used to parallelize processing a chain by setting 
		up multiple readers starting from different snapshots.

		This flag will be ignored in case there is already a snapshot with the same name available in the data directory,
		that is to prevent replaying the snapshot multiple times in case the node manager is being restarted.

		Note that restoring from snapshot means the blocks and state directories are removed before replaying!

		The store url accepts all protocols supported by dstore, which currently are file://, s3://, gs:// and az://. 
		For more infos about supported urls see: https://github.com/streamingfast/dstore
	`))
	//cmd.Flags().Bool(flagPrefix+"bootstrap-keep-blocks", false, FlagDescription(`
	//	If this flag is set the bootstrapping process will not clean the blocks directory before replaying the chain
	//	from snapshot. This is only taking affect if reader-node-bootstrap-snapshot-url is set.
	//
	//	The recommended nodeos config for reader-nodes is to not create a blocks.log in the first place, which can be
	//	set using 'block-log-retain-blocks = 0' in the config.ini.
	//`))
	cmd.Flags().String(flagPrefix+"nodeos-api-addr", NodeosAPIAddr, "Target API address to communicate with underlying nodeos")
}

func registerNode(kind string, extraFlagRegistration func(cmd *cobra.Command) error, managerAPIaddr string) {
	if kind != "reader" {
		panic(fmt.Errorf("invalid kind value, must be either 'reader', got %q", kind))
	}

	app := fmt.Sprintf("%s-node", kind)
	flagPrefix := fmt.Sprintf("%s-", app)

	launcher.RegisterApp(rootLog, &launcher.AppDef{
		ID:          app,
		Title:       fmt.Sprintf("Acme Node (%s)", kind),
		Description: fmt.Sprintf("Acme %s node with built-in operational manager", kind),
		RegisterFlags: func(cmd *cobra.Command) error {
			registerCommonNodeFlags(cmd, flagPrefix, managerAPIaddr)
			extraFlagRegistration(cmd)
			return nil
		},
		InitFunc: func(runtime *launcher.Runtime) error {
			return nil
		},
		FactoryFunc: nodeFactoryFunc(flagPrefix, kind),
	})
}

func nodeFactoryFunc(flagPrefix, kind string) func(*launcher.Runtime) (launcher.App, error) {
	return func(runtime *launcher.Runtime) (launcher.App, error) {
		var appLogger *zap.Logger
		var appTracer logging.Tracer
		var supervisedProcessLogger *zap.Logger

		switch kind {
		case "reader":
			appLogger = readerLogger
			appTracer = readerTracer
			supervisedProcessLogger = readerAcmeChainLogger
		default:
			panic(fmt.Errorf("unknown node kind %q", kind))
		}

		sfDataDir := runtime.AbsDataDir

		nodePath := viper.GetString(flagPrefix + "path")
		nodeDataDir := replaceNodeRole(kind, mustReplaceDataDir(sfDataDir, viper.GetString(flagPrefix+"data-dir")))
		nodeConfigDir := replaceNodeRole(kind, mustReplaceDataDir(sfDataDir, viper.GetString(flagPrefix+"config-dir")))

		readinessMaxLatency := viper.GetDuration(flagPrefix + "readiness-max-latency")
		debugFirehose := viper.GetBool(flagPrefix + "debug-firehose-logs")
		logToZap := viper.GetBool(flagPrefix + "log-to-zap")
		shutdownDelay := viper.GetDuration("common-system-shutdown-signal-delay") // we reuse this global value
		httpAddr := viper.GetString(flagPrefix + "manager-api-addr")

		arguments := viper.GetString(flagPrefix + "arguments")
		nodeArguments, err := buildNodeArguments(
			nodeDataDir,
			kind,
			arguments,
		)
		if err != nil {
			return nil, fmt.Errorf("cannot build node bootstrap arguments")
		}
		metricsAndReadinessManager := buildMetricsAndReadinessManager(flagPrefix, readinessMaxLatency)

		supervisor, err := nodemanager.NewSuperviser(
			debugFirehose,
			metricsAndReadinessManager.UpdateHeadBlock,
			&nodemanager.NodeosOptions{
				LocalNodeEndpoint:    viper.GetString(flagPrefix + "nodeos-api-addr"),
				ConfigDir:            nodeConfigDir,
				BinPath:              nodePath,
				DataDir:              nodeDataDir,
				BootstrapSnapshotUrl: viper.GetString(flagPrefix + "bootstrap-snapshot-url"),
				AdditionalArgs:       nodeArguments,
				LogToZap:             logToZap,
			},
			appLogger,
			supervisedProcessLogger,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize supervisor: %w", err)
		}

		chainOperator, err := operator.New(
			appLogger,
			supervisor,
			metricsAndReadinessManager,
			&operator.Options{
				Bootstrapper:               supervisor,
				EnableSupervisorMonitoring: true,
				ShutdownDelay:              shutdownDelay,
			})
		if err != nil {
			return nil, fmt.Errorf("unable to create chain operator: %w", err)
		}

		blockStreamServer := blockstream.NewUnmanagedServer(blockstream.ServerOptionWithLogger(appLogger))
		oneBlocksStoreURL := mustReplaceDataDir(sfDataDir, viper.GetString("common-one-block-store-url"))
		workingDir := mustReplaceDataDir(sfDataDir, viper.GetString("reader-node-working-dir"))
		gprcListenAddr := viper.GetString("reader-node-grpc-listen-addr")
		batchStartBlockNum := viper.GetUint64("reader-node-start-block-num")
		batchStopBlockNum := viper.GetUint64("reader-node-stop-block-num")
		oneBlockFileSuffix := viper.GetString("reader-node-one-block-suffix")
		blocksChanCapacity := viper.GetInt("reader-node-blocks-chan-capacity")

		readerPlugin, err := reader.NewMindReaderPlugin(
			oneBlocksStoreURL,
			workingDir,
			func(lines chan string) (reader.ConsolerReader, error) {
				return codec.NewConsoleReader(appLogger, lines)
			},
			batchStartBlockNum,
			batchStopBlockNum,
			blocksChanCapacity,
			metricsAndReadinessManager.UpdateHeadBlock,
			func(error) {
				chainOperator.Shutdown(nil)
			},
			oneBlockFileSuffix,
			blockStreamServer,
			appLogger,
			appTracer,
		)
		if err != nil {
			return nil, fmt.Errorf("new reader plugin: %w", err)
		}

		supervisor.RegisterLogPlugin(readerPlugin)

		return nodeManagerApp.New(&nodeManagerApp.Config{
			HTTPAddr: httpAddr,
			GRPCAddr: gprcListenAddr,
		}, &nodeManagerApp.Modules{
			Operator:                   chainOperator,
			MindreaderPlugin:           readerPlugin,
			MetricsAndReadinessManager: metricsAndReadinessManager,
			RegisterGRPCService: func(server grpc.ServiceRegistrar) error {
				pbheadinfo.RegisterHeadInfoServer(server, blockStreamServer)
				pbbstream.RegisterBlockStreamServer(server, blockStreamServer)

				return nil
			},
		}, appLogger), nil
	}
}

type nodeArgsByRole map[string]string

func buildNodeArguments(nodeDataDir, nodeRole string, args string) ([]string, error) {

	// todo figure out roles here, might be reader, producer, bootstrap, ... ?

	typeRoles := nodeArgsByRole{
		"reader": "{extra-arg}",
	}

	argsString, ok := typeRoles[nodeRole]
	if !ok {
		return nil, fmt.Errorf("invalid node role: %s", nodeRole)
	}

	if strings.HasPrefix(args, "+") {
		argsString = strings.Replace(argsString, "{extra-arg}", args[1:], -1)
	} else if args == "" {
		argsString = strings.Replace(argsString, "{extra-arg}", "", -1)
	} else {
		argsString = args
	}

	argsString = strings.Replace(argsString, "{node-data-dir}", nodeDataDir, -1)
	fmt.Println(argsString)
	argsSlice := strings.Fields(argsString)
	return argsSlice, nil
}

func buildMetricsAndReadinessManager(name string, maxLatency time.Duration) *nodeManager.MetricsAndReadinessManager {
	headBlockTimeDrift := metrics.NewHeadBlockTimeDrift(name)
	headBlockNumber := metrics.NewHeadBlockNumber(name)
	appReadiness := metrics.NewAppReadiness(name)

	metricsAndReadinessManager := nodeManager.NewMetricsAndReadinessManager(
		headBlockTimeDrift,
		headBlockNumber,
		appReadiness,
		maxLatency,
	)
	return metricsAndReadinessManager
}

func replaceNodeRole(nodeRole, in string) string {
	return strings.Replace(in, "{node-role}", nodeRole, -1)
}
