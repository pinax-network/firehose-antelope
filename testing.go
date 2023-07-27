package fireantelope

import (
	pbantelope "github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/streamingfast/bstream"
	"google.golang.org/protobuf/proto"
)

func TestingInitBstream() {
	// Should be aligned with firecore. Chain as defined in `cmd/fireantelope/main.go`
	bstream.InitGeneric("EOS", 1, func() proto.Message {
		return new(pbantelope.Block)
	})
}
