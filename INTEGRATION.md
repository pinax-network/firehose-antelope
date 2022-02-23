# Chain Integration Document

We have built an end-to-end template, to start the on-boarding process of new chains. 

# Running the Acme



Acme -> Chain name
firehose-acme > fire
// script names in /Users/julien/codebase/sf/firehose-acme/tools/firedummy/scripts
"github.com/streamingfast/firehose-acme/pb/sf/dummy/codec/v1"

## Release

Use the `./bin/release.sh` Bash script to perform a new release. It will ask you questions
as well as driving all the required commands, performing the necessary operation automatically.
The Bash script runs in dry-mode by default, so you can check first that everything is all right.

Releases are performed using [goreleaser](https://goreleaser.com/).

## Contributing

**Issues and PR in this repo related strictly to the Dummy on StreamingFast.**

Report any protocol-specific issues in their
[respective repositories](https://github.com/streamingfast/streamingfast#protocols)

**Please first refer to the general
[StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md)**,
if you wish to contribute to this code base.

This codebase uses unit tests extensively, please write and run tests.

## License

[Apache 2.0](LICENSE)
