package tests

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


import (
	"encoding/binary"
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

func (s Stringy) SignableBytes() []byte {
	return []byte(s.S)
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

func (i Inty) SignableBytes() []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(i.I))
	return bytes
}

var Tmap = map[tx.TxID]tx.Transactable{
	tx.TxID(1): &Stringy{},
	tx.TxID(2): &Inty{},
}
