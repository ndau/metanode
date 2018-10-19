// This file contains consensus connection methods for the App

package app

import (
	"fmt"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// InitChain performs necessary chain initialization.
//
// This includes saving the initial validator set in the local state.
func (app *App) InitChain(req abci.RequestInitChain) (response abci.ResponseInitChain) {
	logger := app.logRequestBare("InitChain", nil)

	// now add the initial validators set
	for _, v := range req.Validators {
		app.state.UpdateValidator(app.db, v)
	}

	// commiting here ensures two things:
	// 1. we actually have a head value
	// 2. the initial validators are present from tendermint height 0
	app.SetHeight(app.height + 1)
	err := app.commit(logger)
	if err != nil {
		logger.WithError(err).Error("InitChain app commit failed")
		// fail fast if we can't actually initialize the chain
		panic(err.Error())
	}

	app.ValUpdates = make([]abci.ValidatorUpdate, 0)
	return
}

// BeginBlock tracks the block hash and header information
func (app *App) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	var logger log.FieldLogger
	logger = app.DecoratedLogger().WithFields(log.Fields{
		"tm.height": req.GetHeader().Height,
		"tm.time":   req.GetHeader().Time,
		"tm.hash":   fmt.Sprintf("%x", req.GetHash()),
	})
	logger = app.logRequest("BeginBlock", logger)
	// reset valset changes
	app.ValUpdates = make([]abci.ValidatorUpdate, 0)
	app.SetHeight(uint64(req.GetHeader().Height))
	return abci.ResponseBeginBlock{}
}

// DeliverTx services DeliverTx requests
func (app *App) DeliverTx(bytes []byte) (response abci.ResponseDeliverTx) {
	tx, rc, logger, err := app.validateTransactable(bytes)
	app.logRequest("DeliverTx", logger)
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
func (app *App) EndBlock(req abci.RequestEndBlock) abci.ResponseEndBlock {
	app.logRequest("EndBlock", nil)
	return abci.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

// Commit saves a new version
//
// Panics if InitChain has not been called.
func (app *App) Commit() abci.ResponseCommit {
	var logger log.FieldLogger
	logger = app.GetLogger().WithField("qty transactions in block", app.transactionsPending)
	logger = app.logRequest("Commit", logger)

	if app.transactionsPending > 0 {
		app.transactionsPending = 0
		err := app.commit(logger)
		if err != nil {
			logger.WithError(err).Error("Failed to commit block")
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

	return abci.ResponseCommit{Data: app.Hash()}
}
