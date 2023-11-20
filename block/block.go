package block

import (
	"time"

	ts "github.com/PyMarcus/blockchain/transaction"
)

type Block struct {
	Index        int              `json:"index"`
	Timestamp    time.Time        `json:"timestamp"`
	Transactions []ts.Transaction `json:"transactions"`
	Proof        int64			  `json:"proof"`
	PreviousHash string			  `json:"previous_hash"`
}
