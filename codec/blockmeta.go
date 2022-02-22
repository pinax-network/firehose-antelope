package codec

import (
	"container/heap"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/streamingfast/near-go/rpc"
)

type blockMeta struct {
	id        string
	number    uint64
	blockTime time.Time
}

func (m *blockMeta) String() string {
	return fmt.Sprintf("#%d (%s @ %s)", m.number, m.id, m.blockTime)
}

type blockMetas []*blockMeta

func (ms blockMetas) String() string {
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = m.String()
	}

	return strings.Join(out, ", ")
}

type blockMetaGetter interface {
	getBlockMeta(id string) (*blockMeta, error)
}

type blockMetaGetterFunc func(id string) (*blockMeta, error)

func (f blockMetaGetterFunc) getBlockMeta(id string) (*blockMeta, error) {
	return f(id)
}

type RPCBlockMetaGetter struct {
	client *rpc.Client
}

func NewRPCBlockMetaGetter(endpointURL string) *RPCBlockMetaGetter {
	return &RPCBlockMetaGetter{client: rpc.NewClient(endpointURL)}
}

func (g *RPCBlockMetaGetter) getBlockMeta(id string) (*blockMeta, error) {
	res, err := g.client.GetBlock(context.Background(), id)
	if err != nil {
		return nil, err
	}

	return &blockMeta{
		id:        res.Header.Hash,
		number:    uint64(res.Header.Height),
		blockTime: time.Unix(res.Header.Timestamp, 0),
	}, nil
}

type blockMetaHeap struct {
	metas      []*blockMeta
	metasIndex map[string]*blockMeta
	getter     blockMetaGetter
}

func newBlockMetaHeap(getter blockMetaGetter) *blockMetaHeap {
	h := &blockMetaHeap{
		metas:      []*blockMeta{},
		metasIndex: map[string]*blockMeta{},
		getter:     getter,
	}
	return h
}

func (h *blockMetaHeap) get(id string) (*blockMeta, error) {
	if bm, ok := h.metasIndex[id]; ok {
		return bm, nil
	}

	bm, err := h.getter.getBlockMeta(id)
	if err != nil {
		return nil, fmt.Errorf("getting block for id: %s, %w", id, err)
	}

	if bm == nil {
		return nil, fmt.Errorf("block getter return nil block for id: %s", id)
	}

	heap.Push(h, bm)
	return bm, nil
}

func (h *blockMetaHeap) Len() int {
	return len(h.metas)
}

func (h *blockMetaHeap) Less(i, j int) bool {
	return h.metas[i].blockTime.Before(h.metas[j].blockTime)
}

func (h *blockMetaHeap) Swap(i, j int) {
	if len(h.metas) == 0 {
		return
	}

	h.metas[i], h.metas[j] = h.metas[j], h.metas[i]
}

func (h *blockMetaHeap) Push(x interface{}) {
	bm := x.(*blockMeta)
	h.metas = append(h.metas, bm)
	h.metasIndex[bm.id] = bm
}

func (h *blockMetaHeap) Pop() interface{} {
	old := h.metas
	if len(old) == 0 {
		return nil
	}
	n := len(old)
	bm := old[n-1]
	h.metas = old[0 : n-1]
	delete(h.metasIndex, bm.id)
	return bm
}
