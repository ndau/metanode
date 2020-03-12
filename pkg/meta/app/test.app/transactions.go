package testapp

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

	metatx "github.com/ndau/metanode/pkg/meta/transaction"
)

//go:generate msgp -tests=0

// TxIDs is the testapp transactions
var TxIDs = metatx.TxIDMap{
	metatx.TxID(1): &Add{},
}

// Add transactions add an appropriate amount to the state
type Add struct {
	Qty int
}

var _ metatx.Transactable = (*Add)(nil)

// Validate implements Transactable
func (a Add) Validate(interface{}) error {
	if a.Qty < 0 {
		return fmt.Errorf("Adding negatives is silly")
	}
	return nil
}

// Apply implements Transactable
func (a Add) Apply(appI interface{}) error {
	app := appI.(*TestApp)
	return app.UpdateCount(func(c *uint64) error {
		*c += uint64(a.Qty)
		return nil
	})
}

// SignableBytes implements Transactable
func (a Add) SignableBytes() []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(a.Qty))
	return bytes
}
