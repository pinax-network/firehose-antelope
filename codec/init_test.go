package codec

import "github.com/streamingfast/logging"

var zlogTest, _ = logging.PackageLogger("fireantelope", "github.com/EOS-Nation/firehose-antelope/codec.tests")

func init() {
	logging.InstantiateLoggers()
}

type ObjectReader func() (interface{}, error)
