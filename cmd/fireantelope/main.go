package main

import (
	"github.com/pinax-network/firehose-antelope/codec"
	pbantelope "github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	firecore "github.com/streamingfast/firehose-core"
	fhCmd "github.com/streamingfast/firehose-core/cmd"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

func main() {
	fhCmd.Main(Chain())
}

var chain *firecore.Chain[*pbantelope.Block]

func Chain() *firecore.Chain[*pbantelope.Block] {
	if chain != nil {
		return chain
	}

	chain = &firecore.Chain[*pbantelope.Block]{
		ShortName:            "antelope",
		LongName:             "Antelope",
		ExecutableName:       "nodeos",
		FullyQualifiedModule: "github.com/pinax-network/firehose-antelope",
		Version:              version,

		FirstStreamableBlock: 2,

		BlockFactory: func() firecore.Block { return new(pbantelope.Block) },

		ConsoleReaderFactory: codec.NewConsoleReader,

		RegisterExtraStartFlags: func(flags *pflag.FlagSet) {
			flags.String("reader-node-config-file", "", "Node configuration file, the file is copied inside the {data-dir}/reader/data folder Use {hostname} label to use short hostname in path")
			flags.String("reader-node-genesis-file", "./genesis.json", "Node genesis file, the file is copied inside the {data-dir}/reader/data folder. Use {hostname} label to use short hostname in path")
			flags.String("reader-node-key-file", "./node_key.json", "Node key configuration file, the file is copied inside the {data-dir}/reader/data folder. Use {hostname} label to use with short hostname in path")
			flags.Bool("reader-node-overwrite-node-files", false, "Force download of node-key and config files even if they already exist on the machine.")
		},

		Tools: &firecore.ToolsConfig[*pbantelope.Block]{

			RegisterExtraCmd: func(chain *firecore.Chain[*pbantelope.Block], parent *cobra.Command, zlog *zap.Logger, tracer logging.Tracer) error {
				parent.AddCommand(newCheckBlocksCmd(zlog))

				return nil
			},

			SanitizeBlockForCompare: sanitizeBlockForCompare,
		},
	}

	return chain
}

// Version value, injected via go build `ldflags` at build time
var version = "dev"
