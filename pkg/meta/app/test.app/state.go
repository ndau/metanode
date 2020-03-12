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
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	metast "github.com/ndau/metanode/pkg/meta/state"
	util "github.com/ndau/noms-util"
	"github.com/pkg/errors"
)

// TestState is a super simple test state
type TestState struct {
	Number util.Int
}

var _ metast.State = (*TestState)(nil)

// MarshalNoms implements metast.State
func (t TestState) MarshalNoms(vrw nt.ValueReadWriter) (nt.Value, error) {
	numValue, err := t.Number.MarshalNoms(vrw)
	if err != nil {
		return nil, err
	}
	return marshal.Marshal(vrw, nt.NewStruct("TestState", nt.StructData{
		"Number": numValue,
	}))
}

// UnmarshalNoms implements metast.State
func (t *TestState) UnmarshalNoms(v nt.Value) (err error) {
	strct, isStruct := v.(nt.Struct)
	if !isStruct {
		return errors.New("TestState.UnmarshalNoms: v is not a struct")
	}
	numVal, hasNumVal := strct.MaybeGet("Number")
	if !hasNumVal {
		return errors.New("TestState.UnmarshalNoms: Number not found")
	}
	return errors.Wrap(t.Number.UnmarshalNoms(numVal), "TestState.UnmarshalNoms")
}

// Init satisfies metast.State
func (*TestState) Init(vrw nt.ValueReadWriter) {}
