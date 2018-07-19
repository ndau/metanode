package metatx

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

//go:generate msgp
//msgp:shim TxID as:uint8 using:uint8/TxID
//msgp:ignore TxIDMap

// A TxID is a unique identity for this transaction type.
//
// This eases disambiguation for deserialization.
type TxID uint8

// A TxIDMap is a map of unique IDs to Transactable example objects
//
// It should never be mutated by code.
//
// Example objects should be empty objects of the relevant Transactable type.
type TxIDMap map[TxID]Transactable

// Get the ID associated with a Transactable type
//
// We're stuck with linear scan here, which isn't ideal, but for the
// size of transactables we expect, the penalty shouldn't be too bad.
func txIDOf(txab Transactable, idMap TxIDMap) (TxID, error) {
	txabName := reflect.TypeOf(txab).Elem().Name()
	if len(txabName) == 0 {
		return 0, errors.New("anonymous types are not Transactable")
	}
	for id, example := range idMap {
		if reflect.TypeOf(example).Elem().Name() == txabName {
			return id, nil
		}
	}
	return 0, fmt.Errorf("Supplied type `%s` not in `idMap`", txabName)
}
