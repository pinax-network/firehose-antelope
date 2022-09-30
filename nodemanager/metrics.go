package nodemanager

import (
	"github.com/streamingfast/dmetrics"
)

var metricSet = dmetrics.NewSet()

var (
	leapConnectedPeers = metricSet.NewGauge("nodeos_connected_peers_total", "Number of connected peers")
	leapDbSizeInfo     = metricSet.NewGaugeVec("nodeos_db_size_info_bytes", []string{"metric"}, "DB size from Nodeos")
)
