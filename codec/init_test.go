package codec

import (
	"github.com/streamingfast/logging"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var zlogTest, _ = logging.PackageLogger("fireantelope", "github.com/EOS-Nation/firehose-antelope/codec.tests")

func init() {
	logging.InstantiateLoggers()
}

type ObjectReader func() (interface{}, error)

func MarshalIndentToString(m proto.Message, indent string) (string, error) {
	res, err := protojson.MarshalOptions{Indent: indent}.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(res), err
}