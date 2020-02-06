package metatx

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


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

// NameOf uses reflection to get the name associated with a Transactable
func NameOf(txab Transactable) string {
	return reflect.TypeOf(txab).Elem().Name()
}

// TxIDOf uses reflection to get the ID associated with a Transactable type.
//
// We're stuck with linear scan here, which isn't ideal, but for the
// size of transactables we expect, the penalty shouldn't be too bad.
func TxIDOf(txab Transactable, idMap TxIDMap) (TxID, error) {
	txabName := NameOf(txab)
	if len(txabName) == 0 {
		return 0, errors.New("anonymous types are not Transactable")
	}
	for id, example := range idMap {
		if NameOf(example) == txabName {
			return id, nil
		}
	}
	return 0, fmt.Errorf("Supplied type `%s` not in `idMap`", txabName)
}
