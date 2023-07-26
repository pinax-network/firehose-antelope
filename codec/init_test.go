package codec

import (
	fireantelope "github.com/pinax-network/firehose-antelope"
	"github.com/streamingfast/logging"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var zlogTest, _ = logging.PackageLogger("fireantelope", "github.com/pinax-network/firehose-antelope/codec.tests")

func init() {
	logging.InstantiateLoggers()
	fireantelope.TestingInitBstream()
}

type ObjectReader func() (interface{}, error)

func MarshalIndentToString(m proto.Message, indent string) (string, error) {
	res, err := protojson.MarshalOptions{Indent: indent}.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(res), err
}
