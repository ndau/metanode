// This file contains consensus connection methods for the App

package app

import (
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/tendermint/tendermint/abci/types"
)

// InitChain performs necessary chain initialization.
//
// This includes saving the initial validator set in the local state.
func (app *App) InitChain(req types.RequestInitChain) (response types.ResponseInitChain) {
	logger := app.logRequestBare("InitChain")

	// now add the initial validators set
	for _, v := range req.Validators {
		app.state.UpdateValidator(app.db, v)
	}

	// commiting here ensures two things:
	// 1. we actually have a head value
	// 2. the initial validators are present from tendermint height 0
	app.SetHeight(app.height + 1)
	err := app.commit()
	if err != nil {
		logger.Error(err.Error())
		// fail fast if we can't actually initialize the chain
		panic(err.Error())
	}

	app.ValUpdates = make([]types.Validator, 0)
	return
}

// BeginBlock tracks the block hash and header information
func (app *App) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.logRequest("BeginBlock")
	// reset valset changes
	app.ValUpdates = make([]types.Validator, 0)
	app.SetHeight(uint64(req.GetHeader().Height))
	return types.ResponseBeginBlock{}
}

// DeliverTx services DeliverTx requests
func (app *App) DeliverTx(bytes []byte) (response types.ResponseDeliverTx) {
	app.logRequest("DeliverTx")
	tx, rc, err := app.validateTransactable(bytes)
	response.Code = rc
	if err != nil {
		response.Log = err.Error()
		return
	}
	app.checkChild()
	err = tx.Apply(app.childApp)
	if err == nil {
		app.transactionsPending++
	} else {
		response.Code = uint32(code.ErrorApplyingTransaction)
		response.Log = err.Error()
	}
	return
}

// EndBlock updates the validator set
func (app *App) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	app.logRequest("EndBlock")
	return types.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

// Commit saves a new version
//
// Panics if InitChain has not been called.
func (app *App) Commit() types.ResponseCommit {
	logger := app.logRequest("Commit").WithField("qty transactions in block", app.transactionsPending)

	if app.transactionsPending > 0 {
		app.transactionsPending = 0
		err := app.commit()
		if err != nil {
			logger.Error("Failed to commit block")
			// A panic is appropriate here because the one thing we do _not_ want
			// in the event that a block cannot be committed is for the app to
			// just keep ticking along as if things were ok. Crashing the
			// app should kill the whole node service, which in turn should
			// give human operators a chance to figure out what went wrong.
			//
			// There is no noms documentation stating what kind of errors can
			// be expected from this, but we'd expect them to be mostly I/O
			// issues. In that case, restarting the service, potentially
			// automatically, and recovering state from the rest of the chain
			// is the best way forward.
			panic(err)
		}

		logger.Info("Committed noms block")
	} else {
		logger.Info("Skipped noms commit")
	}

	return types.ResponseCommit{Data: app.Hash()}
}
