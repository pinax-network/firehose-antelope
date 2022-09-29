package types

import (
	"fmt"

	"github.com/streamingfast/bstream"
	pbantelope "github.com/streamingfast/firehose-acme/types/pb/sf/antelope/type/v1"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	"google.golang.org/protobuf/proto"
)

func BlockFromProto(b *pbantelope.Block) (*bstream.Block, error) {
	content, err := proto.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal to binary form: %s", err)
	}

	block := &bstream.Block{
		Id:             b.GetId(),
		Number:         uint64(b.GetNumber()),
		PreviousId:     b.Header.Previous,
		Timestamp:      b.Header.Timestamp.AsTime(),
		LibNum:         uint64(b.GetDposIrreversibleBlocknum()),
		PayloadKind:    pbbstream.Protocol_EOS,
		PayloadVersion: 1,
	}

	return bstream.GetBlockPayloadSetter(block, content)
}
