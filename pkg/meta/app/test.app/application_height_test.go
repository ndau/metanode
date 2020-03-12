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
	"math/rand"
	"testing"
	"time"

	"github.com/ndau/metanode/pkg/meta/app/code"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func issueBlock(t *testing.T, app *TestApp, height uint64, txs ...metatx.Transactable) {
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Height: int64(height),
		Time:   time.Now(),
	}})
	for _, tx := range txs {
		bytes, err := metatx.Marshal(tx, TxIDs)
		require.NoError(t, err)
		resp := app.DeliverTx(abci.RequestDeliverTx{Tx: bytes})
		t.Log(resp.Log)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

// Creating noms blocks only when there are transactions is all well
// and good, but we need to ensure that the app height is always what
// tendermint expects.
func TestAppHeightFollowsTendermint(t *testing.T) {
	app, err := NewTestApp()
	require.NoError(t, err)

	require.Equal(t, uint64(0), app.Height())
	app.InitChain(abci.RequestInitChain{})
	require.Equal(t, uint64(1), app.Height())

	apphash := app.Hash()

	for i := uint64(2); i < 6; i++ {
		issueBlock(t, app, i)
		require.Equal(t, i, app.Height())
	}

	// despite having just sent some blocks, the app hash must not have
	// changed, as there have been no transactions
	require.Equal(t, apphash, app.Hash())
}

// let's test that the height and app hash remain consistent across a few
// iterations of the above
func TestAppHeightAndHashUpdatePerTM(t *testing.T) {
	app, err := NewTestApp()
	require.NoError(t, err)

	require.Equal(t, uint64(0), app.Height())
	app.InitChain(abci.RequestInitChain{})
	require.Equal(t, uint64(1), app.Height())

	tmBlock := uint64(0)
	outerLimit := int(rand.Int31n(10))
	for outer := 0; outer <= outerLimit; outer++ {
		apphash := app.Hash()

		innerLimit := int(rand.Int31n(10))
		for inner := 1; inner <= innerLimit; inner++ {
			issueBlock(t, app, tmBlock)
			require.Equal(t, tmBlock, app.Height())
			tmBlock++
		}

		// despite having just sent some blocks, the app hash must not have
		// changed, as there have been no transactions
		require.Equal(t, apphash, app.Hash())

		// now issue a non-empty block
		issueBlock(t, app, tmBlock, &Add{1})
		require.Equal(t, tmBlock, app.Height())
		tmBlock++

		// issuing a transaction must have updated the actual state and
		// created a new app hash
		require.NotEqual(t, apphash, app.Hash())
	}
}
