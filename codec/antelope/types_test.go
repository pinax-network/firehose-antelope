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

	qcBytes, err := hex.DecodeString("af2d12000115fffb1f00c001999edd942be650f36390b24240a943850ea5afbb4dd5e246320a2569eca245b489b600cf12529bd4f313e8d9f3fe9812910e0ed0092e214dd1042d4caf10fc4ed261a854482bc2a46e32999ef69f07eb8c791a72a6f848cfea9c98f1a72ed500b8c513b133d6cc5efb990b212968e69442d2e063473ad32f9407b978098c18b40aa6ef7fa68dd9b679c16052943e4e0a3f81887c4513ba45e9484b37f4d86b919e401ab6f9d2da679c285f7724f9fdf0a640ac2d9b814a722ccab0943c181a07")
	assert.NoError(t, err)

	res := &eos.QuorumCertificateExtension{}
	err = eos.NewDecoder(qcBytes).Decode(res)
	assert.NoError(t, err)

	fmt.Println(res.QuorumCertificate.ValidQuorumCertificate.BlsAggregateSignature.String())

	resJson, _ := json.Marshal(&res)
	fmt.Println(string(resJson))
}
