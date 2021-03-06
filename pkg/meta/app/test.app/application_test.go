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
	"testing"
	"time"

	"github.com/ndau/metanode/pkg/meta/app/code"
	metast "github.com/ndau/metanode/pkg/meta/state"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestCreateTestApp(t *testing.T) {
	_, err := NewTestApp()
	require.NoError(t, err)
}

func TestNegativeAddTxIsInvalid(t *testing.T) {
	app, err := NewTestApp()
	require.NoError(t, err)

	tx := &Add{Qty: -1}
	txBytes, err := metatx.Marshal(tx, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: txBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestPositiveAddTxIsValid(t *testing.T) {
	app, err := NewTestApp()
	require.NoError(t, err)

	tx := &Add{Qty: 1}
	txBytes, err := metatx.Marshal(tx, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: txBytes})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestAddTxProperlyAffectsState(t *testing.T) {
	app, err := NewTestApp()
	require.NoError(t, err)
	app.UpdateCount(func(c *uint64) error {
		*c = 1234
		return nil
	})

	tx := &Add{Qty: 5}
	txBytes, err := metatx.Marshal(tx, TxIDs)
	require.NoError(t, err)

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Time: time.Now()}})
	resp := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, uint64(1239), app.GetCount())
}

func TestInvalidChildStatePreventsTransactions(t *testing.T) {
	const initial = 1234

	app, err := NewTestApp()
	require.NoError(t, err)
	app.UpdateCount(func(c *uint64) error {
		*c = initial
		return nil
	})

	app.SetStateValidity(errors.New("invalid because I said so"))

	tx := &Add{Qty: 5}
	txBytes, err := metatx.Marshal(tx, TxIDs)
	require.NoError(t, err)

	t.Run("CheckTx", func(t *testing.T) {
		resp := app.CheckTx(abci.RequestCheckTx{Tx: txBytes})
		require.Equal(t, code.InvalidNodeState, code.ReturnCode(resp.Code))
	})

	t.Run("DeliverTx", func(t *testing.T) {
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Time: time.Now()}})
		resp := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()

		require.Equal(t, code.InvalidNodeState, code.ReturnCode(resp.Code))
		require.Equal(t, uint64(initial), app.GetCount())
	})

	t.Run("Query", func(t *testing.T) {
		resp := app.Query(abci.RequestQuery{Path: ValueEndpoint})
		require.Equal(t, code.InvalidNodeState, code.ReturnCode(resp.Code))
	})
}

func TestDeferThunk(t *testing.T) {
	app, err := NewTestApp()
	require.NoError(t, err)
	app.UpdateCount(func(c *uint64) error {
		*c = 1234
		return nil
	})

	tx := &Add{Qty: 1}
	txBytes, err := metatx.Marshal(tx, TxIDs)
	require.NoError(t, err)

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Time: time.Now()}})

	// let's pretend that during processing of this tx, the tx handler code
	// decided that it needed to append a thunk
	app.Defer(func(stI metast.State) metast.State {
		st := stI.(*TestState)
		st.Number = 5432
		return st
	})

	resp := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, uint64(5432), app.GetCount())
}

func TestDeferThunkInvalidTx(t *testing.T) {
	app, err := NewTestApp()
	require.NoError(t, err)
	app.UpdateCount(func(c *uint64) error {
		*c = 1234
		return nil
	})

	// remember, negative adds are invalid
	tx := &Add{Qty: -1}
	txBytes, err := metatx.Marshal(tx, TxIDs)
	require.NoError(t, err)

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Time: time.Now()}})

	// let's pretend that during processing of this tx, the tx handler code
	// decided that it needed to append a thunk
	app.Defer(func(stI metast.State) metast.State {
		st := stI.(*TestState)
		st.Number = 5432
		return st
	})

	resp := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	require.Equal(t, uint64(1234), app.GetCount())
}
