package codec_bak

import (
	"github.com/streamingfast/logging"
)

var zlog, _ = logging.PackageLogger("codec", "github.com/streamingfast/firehose-aptos/nodemanager/codec/tests")

func init() {
	logging.InstantiateLoggers()
}
