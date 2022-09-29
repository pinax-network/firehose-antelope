# Chain Integration Document

## Concepts
Blockchain data extraction occurs by a process `Reader`, and a firehose enabled node. We run an instrumented version of a process (usually a node) to sync the chain referred to as `Firehose`.
The 'Firehose' process instruments the blockchain and outputs logs over the standard output pipe, which is subsequently read and processed by the `Reader` process.
The Reader process will read, and stitch together the output of `Firehose` to create rich blockchain data models, which it will subsequently write to
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
As mentioned above, the `Reader` process consumes the data that is extracted and streamed from `Firehose`. In Actuality the Reader
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

modify the flag `reader-node-path: "dchain"` to point to the path of your `dchain` binary you compiled above

## Starting and testing Firehose

*all subsequent commands are run from the `devel/standard/` directory*

Start `fireacme`
```bash
./start.sh
```

This will launch `fireacme` application. Behind the scenes we are starting 3 sub processes: `reader-node`, `relayer`, `firehose`

*reader-node*

The reader-node is a process that runs and manages the blockchain node Geth. It consumes the blockchain data that is
extracted from our instrumented Geth node. The instrumented Geth node outputs individual block data. The reader-node
process will either write individual block data into separate files called one-block files or merge 100 blocks data
together and write into a file called 100-block file.

This behaviour is configurable with the reader-node-merge-and-store-directly flag. When running the reader-node
process with reader-node-merge-and-store-directly flag enable, we say the reader is running in merged mode”.
When the flag is disabled, we will refer to the reader as running in its normal mode of operation.

In the scenario where the reader-node process stores one-block files. We can run a merger process on the side which
would merge the one-block files into 100-block files. When we are syncing the chain we will run the reader-node process
in merged mode. When we are synced we will run the reader-node in it’s regular mode of operation (storing one-block files)

The one-block files and 100-block files will be store in data-dir/storage/merged-blocks and data-dir/storage/one-blocks respectively.
The naming convention of the file is the number of the first block in the file.

As the instrumented node process outputs blocks, you can see the merged block files in the working dir
```bash
ls -las ./firehose-data/storage/merged-blocks
```

We have also built tools that allow you to introspect block files:

```bash
go install ../../cmd/fireantelope && fireantelope tools print blocks --store ./firehose-data/storage/merged-blocks 100
```

At this point we have `reader-node` process running as well a `relayer` & `firehose` process. Both of these processes work together to provide the Firehose data stream.
Once the firehose process is running, it will be listening on port 18015. At it’s core the firehose is a gRPC stream. We can list the available gRPC service

```bash
grpcurl -plaintext localhost:18015 list
```

We can start streaming blocks with `sf.firehose.v2.Stream` Service:

```bash
grpcurl -plaintext -d '{"start_block_num": 10}' -import-path ./proto -proto sf/acme/type/v1/type.proto localhost:18015 sf.firehose.v2.Stream.Blocks
```

# Using `firehose-acme` as a template

One of the main reason we provide a `firehose-acme` repository is to act as a template element that integrators can use to bootstrap
creating the required Firehose chain specific code.

We purposely used `Acme` (and also `acme` and `ACME`) throughout this repository so that integrators can simply copy everything and perform
a global search/replace of this word and use their chain name instead.

As well as this, there is a few files that requires a renaming. Would will find below the instructions to properly make the search/replace
as well as the list of files that should be renamed.

## Cloning

First step is to clone again `firehose-acme` this time to a dedicated repository that will be the one of your chain:

```
git clone git@github.com:streamingfast/firehose-acme.git firehose-<chain>
```

> Don't forget to change `<chain>` by the name of your exact chain like `aptos` so it would became `firehose-aptos`

Then we are going to remove the `.git` folder to start fresh:

```
cd firehose-<chain>
rm -rf .git
git init
```

While not required, I suggest to create an initial commit so it's easier to revert back if you make a mistake
or delete a wrong file:

```
cd firehose-<chain>
git add -A .
git commit -m "Initial commit"
```

## Renaming

Perform a **case-sensitive** search/replace for the following terms:

- `acme` -> `<chain>`
- `Acme` -> `<Chain>`
- `ACME` -> `<CHAIN>`

> Don't forget to change `<chain>` (and their variants) by the name of your exact chain like `aptos` so it would became `aptos`, `Aptos` and `APTOS` respectively.

### Files

```
git mv ./devel/fireacme ./devel/fireaptos
git mv ./cmd/fireacme ./cmd/fireaptos
git mv ./tools/fireacme/scripts/acme-is-running ./tools/fireacme/scripts/aptos-is-running
git mv ./tools/fireacme/scripts/acme-rpc-head-block ./tools/fireacme/scripts/aptos-rpc-head-block
git mv ./tools/fireacme/scripts/acme-resume ./tools/fireacme/scripts/aptos-resume
git mv ./tools/fireacme/scripts/acme-command ./tools/fireacme/scripts/aptos-command
git mv ./tools/fireacme/scripts/acme-debug-firehose-logs-30s ./tools/fireacme/scripts/aptos-debug-deep-mind-30s
git mv ./tools/fireacme/scripts/acme-maintenance ./tools/fireacme/scripts/aptos-maintenance
git mv ./tools/fireacme ./tools/fireaptos
git mv ./types/pb/sf/acme ./types/pb/sf/aptos
```

### Re-generate Protobuf

Once you have performed the renamed of all 3 terms above and file renames, you should re-generate the Protobuf code:

```
cd firehose-<chain>
./types/pb/generate.sh
```

> You will require `protoc`, `protoc-gen-go` and `protoc-gen-go-grpc`. The former can be installed following https://grpc.io/docs/protoc-installation/, the last two can be installed respectively with `go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.25.0` and `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0`.

### Testing

Once everything is done, normally tests should be all good and everything should compile properly:

```
go test ./...
```

### Commit

If everything is fine at that point, you are ready to commit everything and push

```
git add -A .
git commit -m "Renamed Acme to <Chain>"
git add remote origin <url>
git push
```

## License

[Apache 2.0](LICENSE)
