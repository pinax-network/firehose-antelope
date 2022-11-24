package tools

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"
	"strconv"

	"github.com/EOS-Nation/firehose-antelope/types"
	pbantelope "github.com/EOS-Nation/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/spf13/cobra"
	"github.com/streamingfast/bstream"
	sftools "github.com/streamingfast/sf-tools"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func init() {
	Cmd.AddCommand(DownloadFromFirehoseCmd)
	DownloadFromFirehoseCmd.Flags().StringP("api-token-env-var", "a", "FIREHOSE_API_TOKEN", "Look for a JWT in this environment variable to authenticate against endpoint")
	DownloadFromFirehoseCmd.Flags().BoolP("plaintext", "p", false, "Use plaintext connection to firehose")
	DownloadFromFirehoseCmd.Flags().BoolP("insecure", "k", false, "Skip SSL certificate validation when connecting to firehose")
	DownloadFromFirehoseCmd.Flags().Uint32("block-version", 0, "Overwrite the block version. Use this if you want to indicate the blocks have been updated to a new version.")
}

var DownloadFromFirehoseCmd = &cobra.Command{
	Use:     "download-from-firehose <endpoint> <start> <stop> <destination>",
	Short:   "download blocks from firehose and save them to merged-blocks",
	Args:    cobra.ExactArgs(4),
	RunE:    downloadFromFirehoseE,
	Example: "fireantelope tools download-from-firehose eos.firehose.eosnation.io 1000 2000 ./outputdir",
}

var blockVersion = uint32(0)

func downloadFromFirehoseE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	endpoint := args[0]
	start, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("parsing start block num: %w", err)
	}
	stop, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("parsing stop block num: %w", err)
	}
	destFolder := args[3]

	apiTokenEnvVar := mustGetString(cmd, "api-token-env-var")
	apiToken := os.Getenv(apiTokenEnvVar)

	val, err := cmd.Flags().GetUint32("block-version")
	if err == nil && val > 0 {
		blockVersion = val
		zlog.Info("updating block version", zap.Uint32("block_version", blockVersion))
	}

	plaintext := mustGetBool(cmd, "plaintext")
	insecure := mustGetBool(cmd, "insecure")
	var fixerFunc func(*bstream.Block) (*bstream.Block, error)

	return sftools.DownloadFirehoseBlocks(
		ctx,
		endpoint,
		apiToken,
		insecure,
		plaintext,
		start,
		stop,
		destFolder,
		true,
		decodeAndUpdateBlock,
		fixerFunc,
		zlog,
	)
}

func decodeAndUpdateBlock(in *anypb.Any) (*bstream.Block, error) {
	block := &pbantelope.Block{}
	if err := anypb.UnmarshalTo(in, block, proto.UnmarshalOptions{}); err != nil {
		return nil, fmt.Errorf("unmarshal anypb: %w", err)
	}

	if blockVersion > 0 {
		block.Version = blockVersion
	}

	return types.BlockFromProto(block)
}
