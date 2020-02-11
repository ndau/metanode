package state

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


import (
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
)

// State is a generic application state, backed by noms.
//
// For the most part, actual application implementations will
// use type assertions to their particular state implementation
// to gain access to the full range of state functions for that
// application.
type State interface {
	marshal.Marshaler
	marshal.Unmarshaler

	// Init initializes the state.
	//
	// It is expected to be implemented on a pointer receiver, and
	// initialize maps etc as required.
	Init(vrw nt.ValueReadWriter)
}
