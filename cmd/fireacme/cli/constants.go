package cli

const (
	MetricsListenAddr            string = ":9102"
	MindreaderGRPCAddr           string = ":15010"
	NodeManagerAPIAddr           string = ":15041"
	MindreaderNodeManagerAPIAddr string = ":15009"
	MindreaderNodeP2PPort        string = "30305"
	MindreaderNodeRPCPort        string = "8547"
	MergerServingAddr            string = ":15012"
	RelayerServingAddr           string = ":15011"
	FirehoseGRPCServingAddr      string = ":15042"

	ATMDirectory         string = "{sf-data-dir}/atm"
	FirstStreamableBlock int    = 3

	MergedBlocksStoreURL string = "file://{sf-data-dir}/storage/merged-blocks"
	OneBlockStoreURL     string = "file://{sf-data-dir}/storage/one-blocks"

	//Protocol                     = pbbstream.Protocol_NEAR
	//TrxdbDSN              string = "badger://{sf-data-dir}/storage/trxdb"
	//StateDBDSN            string = "badger://{sf-data-dir}/storage/statedb"
	//IndicesStoreURL       string = "file://{sf-data-dir}/storage/indexes"
	//SnapshotsURL          string = "file://{sf-data-dir}/storage/snapshots"
	//DmeshDSN              string = "local://"
	//DmeshServiceVersion   string = "v1"
	//DefaultChainID        uint32 = 123
	//DefaultNetworkID      uint32 = 123
	//DefaultDfuseNetworkID string = "eth-local"
	//
	//MindreaderNodeManagerAPIAddr   string = ":15009"

	//AbiServingAddr                 string = ":15013"
	//BlockmetaServingAddr           string = ":15014"
	//ArchiveServingAddr             string = ":15015"
	//ArchiveHTTPServingAddr         string = ":15016"
	//LiveServingAddr                string = ":15017"
	//RouterServingAddr              string = ":15018"
	//RouterHTTPServingAddr          string = ":15019"
	//KvdbHTTPServingAddr            string = ":15020"
	//IndexerServingAddr             string = ":15021"
	//IndexerHTTPServingAddr         string = ":15022"
	//DgraphqlHTTPServingAddr        string = ":15023"
	//DgraphqlGRPCServingAddr        string = ":15024"
	//ForkresolverServingAddr        string = ":15027"
	//ForkresolverHTTPServingAddr    string = ":15028"
	//FluxDBServingAddr              string = ":15029"
	//EthqHTTPServingAddr            string = ":15030"
	//DashboardGRPCServingAddr       string = ":15031"
	//TrxStateTrackerGRPCServingAddr string = ":15032"
	//TrxStateTrackerHTTPServingAddr string = ":15033"
	//StateDBGRPCServingAddr         string = ":15035"
	//EVMExecutorGRPCServingAddr     string = ":15036"
	//DWeb3HTTPServingAddr           string = ":15037"
	//TokenMetaServingAddr           string = ":15039"
	//TraderServingAddr              string = ":15038"
	//BlockstreamGRPCServingAddr     string = ":15039"
	//BlockstreamHTTPServingAddr     string = ":15040"
	//NodeManagerAPIAddr             string = ":15041"
	//FirehoseGRPCServingAddr        string = ":15042"
	//DashboardHTTPListenAddr        string = ":8081"
	//APIProxyHTTPListenAddr         string = ":8080"

	//
	//// Geth instance port definitions

	NodeP2PPort string = "30303"
	NodeRPCPort string = "3030"
	NodeRPCAddr string = "http://localhost:3030"

	//devMinerAddress       string = "0x821b55d8abe79bc98f05eb675fdc50dfe796b7ab"
)
