// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codec

import (
	"fmt"

	"github.com/EOS-Nation/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/golang/protobuf/proto"
	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
)

func BlockFromProto(b *pbantelope.Block) (*bstream.Block, error) {

	content, err := proto.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal to binary form: %s", err)
	}

	blockTime, err := b.Time()
	if err != nil {
		return nil, err
	}

	blk := &bstream.Block{
		Id:             b.ID(),
		Number:         b.Num(),
		PreviousId:     b.PreviousID(),
		Timestamp:      blockTime,
		LibNum:         b.LIBNum(),
		PayloadKind:    pbbstream.Protocol_EOS,
		PayloadVersion: 1,
	}

	return bstream.GetBlockPayloadSetter(blk, content)
}

// todo legacy function, bstream.StartBlockResolver has been removed, not sure if and how we need to replace
//func BlockstoreStartBlockResolver(blocksStore dstore.Store) bstream.StartBlockResolver {
//	var errFound = errors.New("found")
//
//	return func(ctx context.Context, targetBlockNum uint64) (uint64, string, error) {
//		var dposLibNum uint32
//		num := uint32(targetBlockNum)
//
//		handlerFunc := bstream.HandlerFunc(func(block *bstream.Block, _ interface{}) error {
//			if block.Number == uint64(num) {
//				dposLibNum = uint32(block.LibNum)
//				return errFound
//			}
//
//			return nil
//		})
//
//		fs := bstream.NewFileSource(blocksStore, targetBlockNum, handlerFunc, zlog, nil)
//
//		go fs.Run()
//
//		select {
//		case <-ctx.Done():
//			fs.Shutdown(context.Canceled)
//			return 0, "", ctx.Err()
//		case <-fs.Terminated():
//		}
//
//		if dposLibNum != 0 {
//			return uint64(dposLibNum), "", nil
//		}
//		return 0, "", fs.Err()
//	}
//}
