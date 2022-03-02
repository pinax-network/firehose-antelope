# Chain Integration Document

## Concepts
Blockchain data extraction occurs by two processes running conjunction a `Deepmind` & `Mindreader`. We run an instrumented version of a process (usually a node) to sync the chain referred to as `DeepMind`.
The DeepMind process instruments the blockchain and outputs logs over the standard output pipe, which is subsequently read and processed by the `Mindreader` process.      
The MindReader process will read, and stitch together the output of `DeepMind` to create rich blockchain data models, which it will subsequently write to
files. The data models in question are [Google Protobuf Structures](https://developers.google.com/protocol-buffers).

#### Data Modeling

Designing the  [Google Protobuf Structures](https://developers.google.com/protocol-buffers) for your given blockcahin is one of the most important steps in an integrators journey.
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

Configure firhose test etup
```bash
cd devel/standard/
vi standard.yaml
```

modify the line `mindreader-node-path: "PATH_TO_DCHAIN"` to have contain the path of your `dchain` binary you compiled above  

## Start a test




## License

[Apache 2.0](LICENSE)
