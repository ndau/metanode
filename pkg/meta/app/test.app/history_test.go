package testapp

// This test file exercises a different module: meta/state/history.go
//
// Unfortunately, placing the test file in that module results in a circular
// import pattern, because we need to import TestApp. Therefore, we've moved this
// test file here.

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

type blockFactory struct {
	app    *TestApp
	t      *testing.T
	height int64
}

func (bf *blockFactory) make(txab ...metatx.Transactable) {
	bf.app.BeginBlock(abci.RequestBeginBlock{
		Header: abci.Header{
			Height: bf.height,
			Time:   time.Now(),
		},
	})

	for _, tx := range txab {
		txb, err := metatx.Marshal(tx, TxIDs)
		require.NoError(bf.t, err)
		dtr := bf.app.DeliverTx(abci.RequestDeliverTx{Tx: txb})
		require.True(bf.t, dtr.IsOK())
	}

	bf.app.EndBlock(abci.RequestEndBlock{Height: bf.height})
	bf.app.Commit()

	require.Equal(bf.t, bf.app.Height(), uint64(bf.height))
	bf.height++
}

func initTest(t *testing.T) (*TestApp, blockFactory) {
	app, err := NewTestApp()
	require.NoError(t, err)

	return app,
		blockFactory{
			app:    app,
			t:      t,
			height: int64(app.Height()),
		}
}

func TestEmptyAppHasNoHistory(t *testing.T) {
	app, _ := initTest(t)

	stateCount := 0
	err := state.IterHistory(app.GetDB(), app.GetDS(), app.GetState(), func(state.State, uint64) error {
		stateCount++
		return nil
	})

	require.NoError(t, err)
	// an empty app should have exactly one state, which is empty
	require.Equal(t, 1, stateCount)
}

func TestHistoryHasCorrectNumberOfStates(t *testing.T) {
	qtyStates := 10

	app, bf := initTest(t)

	// make a bunch of states
	// these can't be empty, because empty states aren't recorded in noms
	for i := 0; i < qtyStates; i++ {
		bf.make(&Add{1})
	}

	// count the states present in the history
	stateCount := 0
	err := state.IterHistory(app.GetDB(), app.GetDS(), app.GetState(), func(state.State, uint64) error {
		stateCount++
		return nil
	})
	require.NoError(t, err)

	// an empty app should have exactly one state, which is empty
	require.Equal(t, 1+qtyStates, stateCount)
}

// create a bunch of states for which we can run tests
func createStates(t *testing.T, app *TestApp, bf *blockFactory) {
	bf.make() // tm height 0
	bf.make() // tm height 1
	bf.make() // tm height 2
	bf.make() // tm height 3
	// set a k-v pair, incrementing the noms height to 3
	bf.make(&Add{4}) // tm height 4
	// don't mess with it for a while
	bf.make() // tm height 5
	bf.make() // tm height 6
	bf.make(&Add{7})
	bf.make(&Add{8})
	require.Equal(t, uint64(8), app.Height())
}

func getExpectedStateAtHeight(height uint64) TestState {
	expect := TestState{}
	for _, eh := range []uint64{4, 7, 8} {
		if height >= eh {
			expect.Number = util.Int(uint64(expect.Number) + eh)
		}
	}
	return expect
}

// Test that we recover correct state at all points in history
func TestHistoryCorrectlyRecoversStates(t *testing.T) {
	app, bf := initTest(t)
	createStates(t, app, &bf)

	// ensure that at all points in history, the app.GetState() is what we expect
	st := app.GetState().(*TestState)
	err := state.IterHistory(app.GetDB(), app.GetDS(), st, func(stateI state.State, height uint64) error {
		t.Log("height:", height)
		expected := getExpectedStateAtHeight(height)
		state := stateI.(*TestState)
		require.Equal(t, &expected, state)

		return nil
	})
	require.NoError(t, err)
}

func TestStateAtHeight(t *testing.T) {
	app, bf := initTest(t)
	createStates(t, app, &bf)

	// height 0 == "current height" in our semantics
	for height := uint64(1); height <= 8; height++ {
		st := TestState{}
		err := state.AtHeight(app.GetDB(), app.GetDS(), &st, height)
		require.NoError(t, err)
		require.Equal(t, getExpectedStateAtHeight(height), st)
	}
}
