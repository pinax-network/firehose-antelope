# Firehose on Antelope

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

This is the Antelope chain-specific implementation part of firehose-core and enables both
[Firehose](https://firehose.streamingfast.io/introduction/firehose-overview)
and [Substreams](https://substreams.streamingfast.io) on Antelope chains with native blocks.

## For Developers

To get started with Firehose and Substreams, you need to sign up on https://app.pinax.network to get yourself an api
key. You'll also find quickstarts there to get you started and all of our available endpoints (we currently provide both
Firehose and Substreams endpoints for EOS, WAX and Telos, as well as different testnets).

For connecting to **Firehose** endpoints, you'll need the Protobufs which are published on
[buf.build](https://buf.build/pinax/firehose-antelope/docs/main). Some Golang example code on how to set up a Firehose
client can be found [here](https://github.com/pinax-network/firehose-examples-go).

To **consume** Antelope Substreams, please have a look at the
[documentation](https://substreams.streamingfast.io/documentation/consume). You can also find Substreams to deploy in
our Substreams repository [here](https://github.com/pinax-network/substreams) and on
[substreams.dev](https://substreams.dev).

To **develop** Antelope Substreams, have a look at
the [documentation](https://substreams.streamingfast.io/documentation/develop) here and at the Pinax SDK for Antelope
Substreams which can be found [here](https://github.com/pinax-network/substreams-antelope).

A collection of resources around Substreams can also be found
on [Awesome Substreams](https://github.com/pinax-network/awesome-substreams).

## For Operators

Please have a look at the documentation [here](https://firehose.streamingfast.io) on how to set up your own Firehose &
Substreams stack. Note that indexing larger Antelope chains such as EOS or WAX requires parallel processing of the chain
and a lot of resources to have the indexing done in a reasonable time frame.

### EOS EVM

This implementation provides native Antelope blocks, including all Antelope specific block data. In case you are looking
for operating Firehose & Substreams for EOS EVM, please have a look at
the [firehose-ethereum](https://github.com/streamingfast/firehose-ethereum) repository; it provides a generic evm poller
to poll the EVM blocks from an RPC node.

## License

[Apache 2.0](LICENSE)
