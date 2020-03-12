package app

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
	"testing"

	nt "github.com/attic-labs/noms/go/types"
	metast "github.com/ndau/metanode/pkg/meta/state"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type TestState struct {
	Value int
}

var _ metast.State = (*TestState)(nil)

func (*TestState) Init(nt.ValueReadWriter) {}

func (ts TestState) MarshalNoms(nt.ValueReadWriter) (nt.Value, error) {
	return nt.Number(ts.Value), nil
}

func (ts *TestState) UnmarshalNoms(v nt.Value) error {
	n, ok := v.(nt.Number)
	if !ok {
		return fmt.Errorf("expected Number, got %T", v)
	}
	ts.Value = int(n)
	return nil
}

func TestUpdateStateOperatesOnACopy(t *testing.T) {
	app, err := NewApp("mem", "test", &TestState{1}, metatx.TxIDMap{})
	require.NoError(t, err)

	s := app.GetState().(*TestState)
	err = app.UpdateState(func(stI metast.State) (metast.State, error) {
		// updater must start with an equal value
		require.Equal(t, s, stI)

		// updater must not change original state yet
		st := stI.(*TestState)
		st.Value = 2
		require.Equal(t, 1, s.Value)

		// returning an error must discard state changes
		return stI, errors.New("must discard changes now")
	})
	require.Error(t, err)
	s2 := app.GetState().(*TestState)
	require.Equal(t, 1, s2.Value)
}

func TestUpdateStateLeakyOperatesOnAReference(t *testing.T) {
	app, err := NewApp("mem", "test", &TestState{1}, metatx.TxIDMap{})
	require.NoError(t, err)

	s := app.GetState().(*TestState)
	err = app.UpdateStateLeaky(func(stI metast.State) (metast.State, error) {
		// updater must start with an equal value
		require.Equal(t, s, stI)

		// updater must change original state
		st := stI.(*TestState)
		st.Value = 2
		require.Equal(t, 2, s.Value)

		// returning an error must discard state changes
		return stI, errors.New("must discard changes now")
	})
	require.Error(t, err)
	s2 := app.GetState().(*TestState)
	require.Equal(t, 2, s2.Value)
}
