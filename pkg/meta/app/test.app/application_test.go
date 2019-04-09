package testapp

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
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

	resp := app.CheckTx(txBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestPositiveAddTxIsValid(t *testing.T) {
	app, err := NewTestApp()
	require.NoError(t, err)

	tx := &Add{Qty: 1}
	txBytes, err := metatx.Marshal(tx, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(txBytes)
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
	resp := app.DeliverTx(txBytes)
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
		resp := app.CheckTx(txBytes)
		require.Equal(t, code.InvalidNodeState, code.ReturnCode(resp.Code))
	})

	t.Run("DeliverTx", func(t *testing.T) {
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Time: time.Now()}})
		resp := app.DeliverTx(txBytes)
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
