package blockfetcher

import (
	"context"
	"time"

	"go.uber.org/zap"

	pbbstream "github.com/streamingfast/bstream/pb/sf/bstream/v1"
	"github.com/streamingfast/eth-go/rpc"
	"github.com/streamingfast/firehose-ethereum/block"
)

type SilkwormBlockFetcher struct {
	fetcher *BlockFetcher
}

func NewSilkwormBlockFetcher(rpcClient *rpc.Client, intervalBetweenFetch time.Duration, latestBlockRetryInterval time.Duration, logger *zap.Logger) *SilkwormBlockFetcher {
	fetcher := NewBlockFetcher(rpcClient, intervalBetweenFetch, latestBlockRetryInterval, block.RpcToEthBlock, logger)
	return &SilkwormBlockFetcher{
		fetcher: fetcher,
	}
}

func (f *SilkwormBlockFetcher) PollingInterval() time.Duration { return 1 * time.Second }

func (f *SilkwormBlockFetcher) Fetch(ctx context.Context, blockNum uint64) (*pbbstream.Block, error) {
	return f.fetcher.Fetch(ctx, blockNum)
}
