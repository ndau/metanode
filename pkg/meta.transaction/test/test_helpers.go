package tests

import (
	"fmt"

	tx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	abci "github.com/tendermint/abci/types"
)

//go:generate msgp -tests=0

// set up some example structs which implement Transactable
// the goal is to ensure that serialization and deserialization
// all work as expected

var _ tx.Transactable = (*Stringy)(nil)

type Stringy struct {
	S string
}

func (Stringy) IsValid(abci.Application) error {
	return nil
}

func (Stringy) Apply(abci.Application) error {
	return nil
}

var _ tx.Transactable = (*Inty)(nil)

type Inty struct {
	I int
}

func (Inty) IsValid(abci.Application) error {
	return fmt.Errorf("Intys are never valid")
}

func (Inty) Apply(abci.Application) error {
	return nil
}

var Tmap = map[tx.TxID]tx.Transactable{
	tx.TxID(1): &Stringy{},
	tx.TxID(2): &Inty{},
}
