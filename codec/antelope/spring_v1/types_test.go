package antelope

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFinalityDataDecoding(t *testing.T) {
	bytes, err := hex.DecodeString(strings.ReplaceAll(finalityDataInput1, "\n", ""))
	require.NoError(t, err)

	finalityData := &FinalityData{}
	require.NoError(t, unmarshalBinary(bytes, finalityData), err)

	finalityJson, err := json.Marshal(finalityData)
	require.NoError(t, err)

	fmt.Println(string(finalityJson))
}

var finalityDataInput1 = `01000000000000000100000050f3c0f34771e2f2b8fcd27547f21fba504fa64708222d41336af632c19b1a887ab43c9132eb17408027d834d2c51e44fd4d51d82a09123f8a728c8a65109872`
