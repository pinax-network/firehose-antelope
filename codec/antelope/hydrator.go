package antelope

import "github.com/EOS-Nation/firehose-antelope/types/pb/sf/antelope/type/v1"

type Hydrator interface {
	// HydrateBlock decodes the received Deep Mind AcceptedBlock data structure against the
	// correct struct for this version of EOSIO supported by this hydrator.
	HydrateBlock(block *pbantelope.Block, input []byte) error

	// DecodeTransactionTrace decodes the received Deep Mind AppliedTransaction data structure against the
	// correct struct for this version of EOSIO supported by this hydrator.
	DecodeTransactionTrace(input []byte, opts ...ConversionOption) (*pbantelope.TransactionTrace, error)
}
