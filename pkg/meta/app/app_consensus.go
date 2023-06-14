// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// This file contains consensus connection methods for the App

package app

import (
	"encoding/hex"
	"fmt"

	"github.com/ndau/metanode/pkg/meta/app/code"
	metast "github.com/ndau/metanode/pkg/meta/state"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	math "github.com/ndau/ndaumath/pkg/types"
	log "github.com/sirupsen/logrus"
	abci "github.com/oneiro-ndev/tendermint.0.32.3/abci/types"
)

// IncrementalIndexer declares methods for incremental indexing.
type IncrementalIndexer interface {
	OnBeginBlock(height uint64, blockTime math.Timestamp, tmHash string) error

	// OnDeliverTx is called only after the tx has been successfully applied
	OnDeliverTx(app interface{}, tx metatx.Transactable) error

	OnCommit() error
}

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

	// Debug - Tendermint 0.33 upgrade: Due to pre-genesis accounting data, the blockchain
	// had to start with preset state data having date which is erlier than genesis date.
	// And so the blockchain has to start from snapshot-1 with the height of 1.
	// Nowadays, new nodes should just need to sync from block zero
	// app.SetHeight(app.height + 1)
	app.SetHeight(app.height)
	// End Debug

	err := app.commit(logger)
	if err != nil {
		logger.WithError(err).Error("InitChain app commit failed")
		// fail fast if we can't actually initialize the chain
		panic(err.Error())
	}

	app.ValUpdates = make([]abci.ValidatorUpdate, 0)

	// Note - Tendermint 0.33 upgrade: Should the app_hash be returned to tendermint to
	// allow starting the block sync? Not with this tendermint version
	// return
	return abci.ResponseInitChain{
		ConsensusParams: req.ConsensusParams,
		Validators:      req.Validators,
	}
	// End Note
}

// BeginBlock tracks the block hash and header information
func (app *App) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	tmHeight := req.GetHeader().Height
	tmTime := req.GetHeader().Time
	tmHash := fmt.Sprintf("%x", req.GetHash())

	var err error
	var logger log.FieldLogger
	logger = app.DecoratedLogger().WithFields(log.Fields{
		"tm.height": tmHeight,
		"tm.time":   tmTime,
		"tm.hash":   tmHash,
	})
	logger = app.requestLogger("BeginBlock", true, logger)

	defer func() {
		if err != nil {
			logger = logger.WithError(err)
			logger.Error("BeginBlock erred")
		} else {
			logger.Info("BeginBlock completed successfully")
		}
	}()

	app.blockTime, err = math.TimestampFrom(tmTime)
	if err != nil {
		// we panic because without a good block time, we can't recover
		// The only error conditions for this function are when the timestamp
		// is before the epoch, and overflowing our time type. Neither case
		// is likely on a running blockchain.
		panic(err)
	}

	app.state.AppendRoundStats(logger, req)

	// reset valset changes
	app.ValUpdates = make([]abci.ValidatorUpdate, 0)
	height := uint64(tmHeight)
	app.SetHeight(height)

	// Tell the search we have a new block on the way.
	search := app.GetSearch()
	if search != nil {
		err = search.OnBeginBlock(height, app.blockTime, tmHash)
		if err != nil {
			logger.WithError(err).Error("Failed to begin block for search")
		}
	}

	return abci.ResponseBeginBlock{}
}

// DeliverTx services DeliverTx requests
func (app *App) DeliverTx(request abci.RequestDeliverTx) (response abci.ResponseDeliverTx) {
	var tx metatx.Transactable
	var err error
	var logger log.FieldLogger

	tx, response.Code, logger, err = app.validateTransactable(request.Tx)

	logger = app.requestLogger("DeliverTx", true, logger)

	defer func() {
		logger = logger.WithField("returnCode", code.ReturnCode(response.Code).String())
		if err != nil {
			logger = logger.WithError(err)
			logger.Error("DeliverTx erred")
		} else {
			logger.Info("DeliverTx completed successfully")
		}

		// no matter if they got applied or not, we don't want to persist any thunks
		// past this tx
		app.deferredThunks = nil
	}()

	if err != nil {
		logger = logger.WithField("err.context", "validating transactable")
		response.Log = err.Error()
		return
	}
	app.checkChild()
	err = tx.Apply(app.childApp)
	if err == nil {
		// wrap the deferred thunks in a format that app.UpdateState can call
		wthunks := make([]func(metast.State) (metast.State, error), 0, len(app.deferredThunks))
		for _, thunk := range app.deferredThunks {
			wthunks = append(wthunks, func(st metast.State) (metast.State, error) {
				st = thunk(st)
				if st == nil {
					// thunks are never allowed to return nil states,
					// and if one does so, we can't recover
					panic("deferred thunk returned nil state")
				}
				return st, nil
			})
		}
		// ignore the returned error: if no thunk errors (and they aren't allowed to!),
		// then the UpdateState call can't error
		app.UpdateState(wthunks...)

		// the qty of pending txs informs whether we noms-commit, or just continue
		app.transactionsPending++

		// Update the search with the new transaction.
		search := app.GetSearch()
		if search != nil {
			err = search.OnDeliverTx(app.childApp, tx)
			if err != nil {
				logger = logger.WithField("err.context", "failed to deliver tx for search")
				response.Code = uint32(code.IndexingError)
				response.Log = err.Error()
			}
		}
	} else {
		logger = logger.WithField("err.context", "applying transaction")
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
	var err error
	var logger log.FieldLogger
	logger = app.DecoratedLogger().WithFields(log.Fields{
		"block.qty_transactions":  app.transactionsPending,
		"abci.sequence":           "start",
		"app.transactionsPending": app.transactionsPending,
		"commit.status":           "pending",
	})
	logger = app.requestLogger("Commit", true, logger)

	finalizeLogger := func() {
		if err != nil {
			logger = logger.WithError(err)
			logger.Error("Commit erred")
		} else {
			logger.Info("Commit completed successfully")
			logger = app.DecoratedLogger().WithFields(log.Fields{
				"Debug.... app hash": hex.EncodeToString(app.Hash()),
			})
		}
	}
	defer finalizeLogger()

	logger = logger.WithField("abci.sequence", "mid")
	if app.transactionsPending > 0 {
		app.transactionsPending = 0
		err = app.commit(logger)
		if err != nil {
			logger = logger.WithFields(log.Fields{
				"err.context":   "failed to commit block",
				"commit.status": "failed",
			})
			// finalize the logger and flush it now; otherwise, the panic here
			// is likely to kill the whole process before it actually gets
			// sent off.
			finalizeLogger()
			type flusher interface {
				Flush()
			}
			lhs := logger.(*log.Entry).Logger.Hooks
			for _, hs := range lhs {
				for _, h := range hs {
					if f, ok := h.(flusher); ok {
						f.Flush()
					}
				}
			}

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

		logger = logger.WithField("commit.status", "success")

		// Index the transactions in the new block.
		search := app.GetSearch()
		if search != nil {
			err = search.OnCommit()
			if err != nil {
				// failing to commit for search doesn't cause tx rejection
				logger.Error("failed to commit for search")
				err = nil
			}
		}
	} else {
		logger = logger.WithField("commit.status", "skipped: no txs pending")
	}
	logger = logger.WithField("abci.sequence", "end")

	return abci.ResponseCommit{Data: app.Hash()}
}
