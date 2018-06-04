package app

import (
	"github.com/attic-labs/noms/go/datas"
	abci "github.com/tendermint/abci/types"
)

// State is a generic application state, backed by noms.
//
// For the most part, actual application implementations will
// use type assertions to their particular state implementation
// to gain access to the full range of state functions for that
// application.
type State interface {
	// Load gets the state from a dataset.
	//
	// If it does not exist, or the dataset has no head, it must be
	// created and appropriately initialized.
	//
	// This is expected to be implemented on a pointer receiver. It
	// will be called on anil pointer of the appropriate type. This
	// function must edit the value of the pointer appropriately to
	// load the state.
	Load(db datas.Database, ds datas.Dataset) (datas.Dataset, error)

	// UpdateValidator updates the state's record of the given validator.
	//
	// When the supplied validator's power is 0, it should be removed.
	UpdateValidator(datas.Database, abci.Validator)

	// Commit commits the current state and returns an updated dataset
	//
	// Note that this is expected to be implemented on a pointer receiver
	// and modify that receiver.
	//
	// If the error returned is not nil, this function should _not_
	// modify the receiver.
	Commit(db datas.Database, ds datas.Dataset) (datas.Dataset, error)
}
