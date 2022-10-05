package transform

import (
	"fmt"
	pbtransform "github.com/EOS-Nation/firehose-antelope/types/pb/sf/antelope/transform/v1"
	"github.com/eoscanada/eos-go"
	"github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Filter struct {
	ActionFilters       []*ActionFilter
	indexStore          dstore.Store
	possibleIndexSizes  []uint64
	sendAllBlockHeaders bool
}

func (f *Filter) String() string {
	//TODO implement me
	panic("implement me")
}

type ActionFilter struct {
	Account eos.AccountName
	Action  eos.AccountName
}

var FilterMessageName = proto.MessageName(&pbtransform.Filter{})

func ActionFilterFactory(indexStore dstore.Store, possibleIndexSizes []uint64) *transform.Factory {
	return &transform.Factory{
		Obj: &pbtransform.Filter{},
		NewFunc: func(message *anypb.Any) (transform.Transform, error) {

			if message.MessageName() != FilterMessageName {
				return nil, fmt.Errorf("expected type url %q, received %q ", FilterMessageName, message.TypeUrl)
			}

			filter := &pbtransform.Filter{}
			err := proto.Unmarshal(message.Value, filter)
			if err != nil {
				return nil, fmt.Errorf("unexpected unmarshall error: %w", err)
			}

			if len(filter.ActionFilters) == 0 && !filter.SendAllBlockHeaders {
				return nil, fmt.Errorf("at least one action filter is required if send_all_block_headers is not set")
			}

			return newFilter(indexStore, possibleIndexSizes, filter.ActionFilters, filter.SendAllBlockHeaders)
		},
	}
}

func newFilter(indexStore dstore.Store, possibleIndexSizes []uint64, pbActionFilters []*pbtransform.ActionFilter, sendAllBlockHeaders bool) (filter *Filter, err error) {

	actionFilters := make([]*ActionFilter, len(pbActionFilters))

	for i, f := range pbActionFilters {
		actionFilters[i] = &ActionFilter{
			Account: eos.AN(eos.NameToString(f.Account)),
			Action:  eos.AN(eos.NameToString(f.Action)),
		}
	}

	return &Filter{
		ActionFilters:       actionFilters,
		indexStore:          indexStore,
		possibleIndexSizes:  possibleIndexSizes,
		sendAllBlockHeaders: sendAllBlockHeaders,
	}, nil
}
