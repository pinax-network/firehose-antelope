# Chain Integration Document

## Concepts
Blockchain data extraction occurs by two processes running conjunction a `Deepmind` & `Mindreader`. We run an instrumented version of a process (usually a node) to sync the chain referred to as `DeepMind`.
The DeepMind process instruments the blockchain and outputs logs over the standard output pipe, which is subsequently read and processed by the `Mindreader` process.
The MindReader process will read, and stitch together the output of `DeepMind` to create rich blockchain data models, which it will subsequently write to
files. The data models in question are [Google Protobuf Structures](https://developers.google.com/protocol-buffers).

#### Data Modeling

Designing the  [Google Protobuf Structures](https://developers.google.com/protocol-buffers) for your given blockchain is one of the most important steps in an integrators journey.
The data structures needs to represent as precisely as possible the on chain data and concepts. By carefully crafting the Protobuf structure, the next steps will be a lot simpler.
The data model need.

As a reference, here is Ethereum's Protobuf Structure:
https://github.com/streamingfast/proto-ethereum/blob/develop/sf/ethereum/codec/v1/codec.proto

# Running the Demo Chain

We have built an end-to-end template, to start the on-boarding process of new chains. This solution consist of:

*firehose-acme*
As mentioned above, the `Mindreader` process consumes the data that is extracted and streamed from `Deeepmind`. In Actuality the MindReader
is one process out of multiple ones that creates the _Firehose_. These processes are launched by one application. This application is
chain specific and by convention, we name is "firehose-<chain-name>". Though this application is chain specific, the structure of the application
is standardized and is quite similar from chain to chain. For convenience, we have create a boiler plate app to help you get started.
We named our chain `Acme` this the app is [firehose-acme](https://github.com/streamingfast/firehose-acme)

*DeepMind*
`Deepmind` consist of an instrumented syncing node. We have created a "dummy" chain to simulate a node process syncing that can be found [https://github.com/streamingfast/dummy-blockchain](https://github.com/streamingfast/dummy-blockchain).

## Setting up the dummy chain

Clone the repository:
```bash
git clone https://github.com/streamingfast/dummy-blockchain.git
cd dummy-blockchain
```

Install dependencies:

```bash
go mod download
```

Then build the binary:
```bash
make build
```

Ensure the build was successful
```bash
./dchain --version
```

Take note of the location of the built `dchain` binary, you will need to configure `firehose-acme` with it.

## Setting up firehose-acme

Clone the repository:

```bash
git clone git@github.com:streamingfast/firehose-acme.git
cd firehose-acme
```

Configure firehose test etup

```bash
cd devel/standard/
vi standard.yaml
```

modify the flag `mindreader-node-path: "dchain"` to point to the path of your `dchain` binary you compiled above

## Starting and testing Firehose

*all subsequent commands are run from the `devel/standard/` directory*

Start `fireacme`
```bash
./start.sh
```

This will launch `fireacme` application. Behind the scenes we are starting 3 sub processes: `mindreader-node`, `relayer`, `firehose`

*mindreader-node*

The mindreader-node is a process that runs and manages the blockchain node Geth. It consumes the blockchain data that is
extracted from our instrumented Geth node. The instrumented Geth node outputs individual block data. The mindreader-node
process will either write individual block data into separate files called one-block files or merge 100 blocks data
together and write into a file called 100-block file.

This behaviour is configurable with the mindreader-node-merge-and-store-directly flag. When running the mindreader-node
process with mindreader-node-merge-and-store-directly flag enable, we say the “mindreader is running in merged mode”.
When the flag is disabled, we will refer to the mindreader as running in its normal mode of operation.

In the scenario where the mindreader-node process stores one-block files. We can run a merger process on the side which
would merge the one-block files into 100-block files. When we are syncing the chain we will run the mindreader-node process
in merged mode. When we are synced we will run the mindreader-node in it’s regular mode of operation (storing one-block files)

The one-block files and 100-block files will be store in data-dir/storage/merged-blocks and data-dir/storage/one-blocks respectively.
The naming convention of the file is the number of the first block in the file.

As the instrumented node process outputs blocks, you can see the merged block files in the working dir
```bash
ls -las ./firehose-data/storage/merged-blocks
```

We have also built tools that allow you to introspect block files:

```bash
go install ../../cmd/fireacme && fireacme tools print blocks --store ./firehose-data/storage/merged-blocks 100
```

At this point we have `mindreader-node` process running as well a `relayer` & `firehose` process. Both of these processes work together to provide the Firehose data stream.
Once the firehose process is running, it will be listening on port 13042. At it’s core the firehose is a gRPC stream. We can list the available gRPC service

```bash
grpcurl -plaintext localhost:17042 list
```

We can start streaming blocks with `sf.firehose.v1.Stream` Service:

```bash
grpcurl -plaintext -d '{"start_block_num": 10}' -import-path ./proto -proto sf/acme/type/v1/type.proto localhost:17042 sf.firehose.v1.Stream.Blocks
```

## License

[Apache 2.0](LICENSE)
