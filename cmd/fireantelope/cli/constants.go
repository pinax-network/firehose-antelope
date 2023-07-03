package cli

const (

	// This should be the block number of the first block that is streamable using Firehose,
	// for example on Ethereum it's set to `0`, the genesis block's number while on EOSIO it's
	// set to 2 (genesis block is 1 there but our instrumentation on this chain instruments
	// only from block #2).
	//
	// This is used in multiple places to determine if we reached the oldest block of the chain.
	FirstStreamableBlock int = 2

	// Should be the number of blocks between two targets before we consider the
	// first as "near" the second. For example if a chain is at block #215 and another
	// source is at block #225, then there is a difference of 10 blocks which is <=
	// than `BlockDifferenceThresholdConsideredNear` which would mean it's "near".
	BlockDifferenceThresholdConsideredNear = 15

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
	SnapshotsURL  string = "file://{sf-data-dir}/storage/snapshots"

	//
	/// Standard Values
	//

	// Common ports
	MetricsListenAddr string = ":9102"

	// Firehose chain specific port
	//
	// The initial 18XXX prefix is different for every chain supported by Firehose.
	// The current prefix is the one you should use for your chain. Once you have copied
	// this whole repository, you should open a PR on firehose-acme to bump it again
	// so the next team supporting Firehose will use 18XXX and so forth.
	ReaderNodeGRPCAddr             string = ":18010"
	ReaderNodeManagerAPIAddr       string = ":18011"
	MergerServingAddr              string = ":18012"
	NodeManagerAPIAddr             string = ":18013"
	RelayerServingAddr             string = ":18014"
	FirehoseGRPCServingAddr        string = ":18015"
	SubstreamsTier1GRPCServingAddr string = ":18044"
	SubstreamsTier2GRPCServingAddr string = ":18045"

	// Data storage default locations
	BlocksCacheDirectory string = "file://{sf-data-dir}/storage/blocks-cache"
	MergedBlocksStoreURL string = "file://{sf-data-dir}/storage/merged-blocks"
	OneBlockStoreURL     string = "file://{sf-data-dir}/storage/one-blocks"
	ForkedBlocksStoreURL string = "file://{sf-data-dir}/storage/forked-blocks"
)
