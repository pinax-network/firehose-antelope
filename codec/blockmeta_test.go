package codec

import (
	"container/heap"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBlockHeap_Push_Get(t *testing.T) {
	now := time.Now()
	blockMetas := []*blockMeta{
		{
			id:        "id.1",
			number:    1,
			blockTime: now.Add(1 * time.Second),
		},
		{
			id:        "id.2",
			number:    2,
			blockTime: now.Add(2 * time.Second),
		},
		{
			id:        "id.3",
			number:    3,
			blockTime: now.Add(3 * time.Second),
		},
	}

	h := newBlockMetaHeap(nil)
	for _, bm := range blockMetas {
		h.Push(bm)
	}

	var err error

	//GET
	bm, err := h.get("id.1")
	require.Equal(t, bm.id, "id.1")
	require.NoError(t, err)

	bm, err = h.get("id.2")
	require.Equal(t, bm.id, "id.2")
	require.NoError(t, err)

	bm, err = h.get("id.3")
	require.Equal(t, bm.id, "id.3")
	require.NoError(t, err)

	//POP
	bm = heap.Pop(h).(*blockMeta)
	require.Equal(t, bm.id, "id.1")

	bm = heap.Pop(h).(*blockMeta)
	require.Equal(t, bm.id, "id.2")

	bm = heap.Pop(h).(*blockMeta)
	require.Equal(t, bm.id, "id.3")
}

func TestBlockHeap_BlockGetter(t *testing.T) {
	getterCallCount := 0
	getter := blockMetaGetterFunc(func(id string) (*blockMeta, error) {
		getterCallCount += 1
		return &blockMeta{
			id:        id,
			number:    1,
			blockTime: time.Now(),
		}, nil
	})

	h := newBlockMetaHeap(getter)

	h.get("id.1")

	require.Equal(t, 1, getterCallCount)

	h.get("id.1")

	require.Equal(t, 1, getterCallCount)

	bm := heap.Pop(h).(*blockMeta)
	require.Equal(t, "id.1", bm.id)
}
