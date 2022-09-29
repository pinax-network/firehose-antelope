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
