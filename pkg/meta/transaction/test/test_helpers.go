package tests

import (
	"fmt"

	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
)

//go:generate msgp -tests=0

// set up some example structs which implement Transactable
// the goal is to ensure that serialization and deserialization
// all work as expected

var _ tx.Transactable = (*Stringy)(nil)

type Stringy struct {
	S string
}

func (Stringy) Validate(interface{}) error {
	return nil
}

func (Stringy) Apply(interface{}) error {
	return nil
}

var _ tx.Transactable = (*Inty)(nil)

type Inty struct {
	I int
}

func (Inty) Validate(interface{}) error {
	return fmt.Errorf("Intys are never valid")
}

func (Inty) Apply(interface{}) error {
	return nil
}

var Tmap = map[tx.TxID]tx.Transactable{
	tx.TxID(1): &Stringy{},
	tx.TxID(2): &Inty{},
}
