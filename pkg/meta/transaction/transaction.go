package metatx

import "github.com/tinylib/msgp/msgp"

//go:generate msgp

// A Transaction is a transaction as recognized by Tendermint
type Transaction struct {
	Nonce          []byte
	TransactableID TxID
	Transactable   msgp.Raw
}
