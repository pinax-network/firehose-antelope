package types

import (
	"fmt"

	pbantelope "github.com/EOS-Nation/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	"google.golang.org/protobuf/proto"
)

func BlockFromProto(b *pbantelope.Block) (*bstream.Block, error) {

	content, err := proto.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal to binary form: %s", err)
	}

	blk := &bstream.Block{
		Id:             b.ID(),
		Number:         b.Num(),
		PreviousId:     b.PreviousID(),
		Timestamp:      b.Time(),
		LibNum:         b.LIBNum(),
		PayloadKind:    pbbstream.Protocol_EOS,
		PayloadVersion: 1,
	}

	return bstream.GetBlockPayloadSetter(blk, content)
}
