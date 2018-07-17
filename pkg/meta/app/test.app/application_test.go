package testapp

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
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

	app.BeginBlock(abci.RequestBeginBlock{})
	resp := app.DeliverTx(txBytes)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, uint64(1239), app.GetCount())
}
