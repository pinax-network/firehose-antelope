package codec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func assertProtoEqual(t *testing.T, expected, actual proto.Message) {
	t.Helper()

	// We use a custom comparison function and than rely on a standard `assert.Equal` so we get some
	// diffing information. Ideally, a better diff would be displayed, good enough for now.
	if !proto.Equal(expected, actual) {
		assert.Equal(t, expected, actual)
	}
}
