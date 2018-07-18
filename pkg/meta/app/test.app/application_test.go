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

func issueBlock(t *testing.T, app *TestApp, height uint64, txs ...metatx.Transactable) {
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Height: int64(height),
	}})
	for _, tx := range txs {
		bytes, err := metatx.Marshal(tx, TxIDs)
		require.NoError(t, err)
		resp := app.DeliverTx(bytes)
		t.Log(resp.Log)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

// Creating noms blocks only when there are transactions is all well
// and good, but we need to ensure that the app height is always what
// tendermint expects.
func TestAppHeightStaysCurrent(t *testing.T) {
	app, err := NewTestApp()
	require.NoError(t, err)

	require.Equal(t, uint64(0), app.Height())
	app.InitChain(abci.RequestInitChain{})
	require.Equal(t, uint64(0), app.Height())

	for i := uint64(1); i < 5; i++ {
		issueBlock(t, app, i)
		require.Equal(t, i, app.Height())
	}

}
