package metatx

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import "github.com/tinylib/msgp/msgp"

//go:generate msgp

// A Transaction is a transaction as recognized by Tendermint
type Transaction struct {
	Nonce          []byte
	TransactableID TxID
	Transactable   msgp.Raw
}
