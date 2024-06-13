package antelope

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/eoscanada/eos-go"
	"testing"
	"unicode/utf8"

	"github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
	"github.com/stretchr/testify/assert"
)

func TestLimitConsoleLengthConversionOption(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		maxByteCount int
		expected     string
	}{
		{"one extra requires truncation, unicode (1 byte)", "abc", 2, "ab"},
		{"exact flush no truncation, unicode (1 byte)", "abc", 3, "abc"},

		{"one extra requires truncation, unicode (multi-byte)", "我我我", 5, "我"},
		{"exact flush no truncation, unicode (multi-byte)", "我我我", 6, "我我"},

		{"truncate before valid multi-byte utf8, nothing before", "🚀", 4, "🚀"},
		{"truncate at 3 before valid multi-byte utf8, nothing before", "🚀", 3, ""},
		{"truncate at 2 before valid multi-byte utf8, nothing before", "🚀", 2, ""},
		{"truncate at 1 before valid multi-byte utf8, nothing before", "🚀", 1, ""},

		{"truncate before valid multi-byte utf8, something before", "我🚀", 7, "我🚀"},
		{"truncate at 3 before valid multi-byte utf8, something before", "我🚀", 6, "我"},
		{"truncate at 2 before valid multi-byte utf8, something before", "我🚀", 5, "我"},
		{"truncate at 1 before valid multi-byte utf8, something before", "我🚀", 4, "我"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actTrace := &pbantelope.ActionTrace{Console: test.in}

			option := LimitConsoleLengthConversionOption(test.maxByteCount)
			option.(ActionConversionOption).Apply(actTrace)

			assert.Equal(t, test.expected, actTrace.Console)
			assert.True(t, utf8.ValidString(actTrace.Console), "The truncated string is not a fully valid utf-8 sequence")
		})
	}
}

func TestBlockExtensionsToDEOS(t *testing.T) {

	qcBytes, err := hex.DecodeString("053512000115fffb1f00c0016d7c4c1762ce7320572ef02a04dcbae1822af1658c2c65b441a6b214e1a7a5fc098667519359e84a19b2a22723905f02cf184696e292af4ac8d35c59818d4f7956a27aad59e802f7f2ead304d32cad13c2570c29e7797b1127a9748740b2b40cb5d1b0ecd83755ce03eaca747228ef1d05316fd970467d1b28529cd0f1f96a4cb750d7460bde5ceb05baa778e3f1fd033175f4689b8745bd89cdaa681883b076e4a5423a9c0e27490e188579be18d1f6932112c389875d9d25f567b06c410418")
	assert.NoError(t, err)

	res := &eos.QuorumCertificateExtension{}
	err = eos.NewDecoder(qcBytes).Decode(res)
	assert.NoError(t, err)

	fmt.Println(res.QuorumCertificate.ValidQuorumCertificate.BlsAggregateSignature.String())

	resJson, _ := json.Marshal(&res)
	fmt.Println(string(resJson))
}
