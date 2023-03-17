package codec

import (
	"github.com/pinax-network/firehose-antelope/types/pb/sf/antelope/type/v1"
)

func DeduplicateTransactionTrace(trx *pbantelope.TransactionTrace) {
	digs := make(map[string]*pbantelope.Action)

	// loop through act_digest
	for _, act := range trx.ActionTraces {
		if act.Receipt != nil {
			act.Receipt.Receiver = ""

			if act.Action.Account == "eosio" && act.Action.Name == "setabi" {
				// In the SLIM case where a `setabi` would make the JSON decoding different
				// after `setabi` is called.  Wow, such powerful!  Execution order assumed.
				digs = make(map[string]*pbantelope.Action)
			}

			d := act.Receipt.Digest
			_, found := digs[d]

			if found {
				act.Action = nil
			} else {
				digs[d] = act.Action
			}
		}

		act.TransactionId = ""
		act.ProducerBlockId = ""
		act.BlockTime = nil
		act.BlockNum = 0
	}
}

func ReduplicateTransactionTrace(trx *pbantelope.TransactionTrace) {
	digs := make(map[string]*pbantelope.Action)
	for _, act := range trx.ActionTraces {
		if act.Receipt != nil {
			if act.Action == nil {
				act.Action = digs[act.Receipt.Digest]
				if act.Action == nil {
					panic("consistency error in deduplicate/reduplicate")
				}
			} else {
				digs[act.Receipt.Digest] = act.Action
			}

			act.Receipt.Receiver = act.Receiver
		}

		act.TransactionId = trx.Id
		act.ProducerBlockId = trx.ProducerBlockId
		act.BlockTime = trx.BlockTime
		act.BlockNum = trx.BlockNum
	}
}
