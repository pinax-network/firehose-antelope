package cli

import pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"

const (
	Protocol = pbbstream.Protocol_EOS

	DefaultChainID      uint32 = 123
	DefaultNetworkID    uint32 = 123
	DefaultDeploymentID string = "antelope-local"

	OneBlockStoreURL     string = "file://{sf-data-dir}/storage/one-blocks"
	ForkedBlocksStoreURL string = "file://{sf-data-dir}/storage/forked-blocks"
	MergedBlocksStoreURL string = "file://{sf-data-dir}/storage/merged-blocks"
	IndexStoreURL        string = "file://{sf-data-dir}/storage/index"
	SnapshotsURL         string = "file://{sf-data-dir}/storage/snapshots"
	StateDBDSN           string = "badger://{sf-data-dir}/storage/statedb"

	MetricsListenAddr string = ":9102"

	BlocksCacheDirectory           string = "{sf-data-dir}/blocks-cache"
	BlockstreamGRPCServingAddr     string = ":18039"
	BlockstreamHTTPServingAddr     string = ":18040"
	EVMExecutorGRPCServingAddr     string = ":18036"
	FirehoseGRPCServingAddr        string = ":18042"
	SubstreamsTier1GRPCServingAddr string = ":18044"
	SubstreamsTier2GRPCServingAddr string = ":18045"
	FirehoseGRPCHealthServingAddr  string = ":18043"
	MergerServingAddr              string = ":18012"
	IndexBuilderServiceAddr        string = ":18043"
	ReaderNodeManagerAPIAddr       string = ":18009"
	ReaderGRPCAddr                 string = ":18010"
	NodeManagerAPIAddr             string = ":18041"
	RelayerServingAddr             string = ":18011"
	TokenMetaServingAddr           string = ":18039"
	TraderServingAddr              string = ":18038"
	StateDBServingAddr             string = ":18029"
	StateDBGRPCServingAddr         string = ":18035"

	CommonAutoMaxProcsFlag              string = "common-auto-max-procs"
	CommonAutoMemLimitFlag              string = "common-auto-mem-limit-percent"
	CommonSystemShutdownSignalDelayFlag string = "common-system-shutdown-signal-delay"

	// Those should be the port the native node is using for P2P and RPC respectively
	// and importantly, they should be different than the `node` ones below. Each chain
	// usually have at least P2P and RPC ports. We suggest to use the standard port on the
	// `node` values below and increment the `reader` ones by 1.
	ReaderNodeP2PPort string = "30304"
	ReaderNodeRPCPort string = "8546"

	// Those should be the port the native node is using for P2P and RPC respectively
	// and importantly, they should be different than the `reader` ones above. We suggest
	// to use the standard port on the `node` values here directly.
	NodeP2PPort string = "30303"
	NodeRPCPort string = "8545"

	// This should be the standard name of the executable that is usually used to
	// sync the chain with the blockchain network. For example on Ethereum where
	// our standard instrumentation if using the Geth client, value is `geth`, on EOSIO
	// chain, it's `nodeos`.
	ChainExecutableName = "nodeos"

	NodeosAPIAddr string = ":8888"
)
