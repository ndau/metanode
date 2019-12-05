// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// This file contains consensus connection methods for the App

package app

import (
	"fmt"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// InitChain performs necessary chain initialization.
//
// This includes saving the initial validator set in the local state.
func (app *App) InitChain(request abci.RequestInitChain) (response abci.ResponseInitChain) {
	var err error
	logger := app.logRequestBare("InitChain", nil)
	defer func() {
		if err != nil {
			logger = logger.WithError(err)
			logger.Error("InitChain erred")
		} else {
			logger.Info("InitChain completed successfully")
		}
	}()

	// now add the initial validators set
	for _, v := range request.Validators {
		app.state.UpdateValidator(app.db, v)
	}

	// commiting here ensures two things:
	// 1. we actually have a head value
	// 2. the initial validators are present from tendermint height 0
	app.SetHeight(app.height + 1)
	err = app.commit(logger)
	if err != nil {
		panic(err.Error())
	}

	app.ValUpdates = make([]abci.ValidatorUpdate, 0)

	if app.indexer != nil {
		start := time.Now()
		err = errors.Wrap(app.indexer.InitChain(request, response, app.GetState()), "indexing")
		duration := time.Since(start)
		logger = logger.WithFields(log.Fields{
			"index.elapsed.ns": duration.Nanoseconds(),
		})
	}

	return
}

// BeginBlock tracks the block hash and header information
func (app *App) BeginBlock(request abci.RequestBeginBlock) (response abci.ResponseBeginBlock) {
	tmHeight := request.GetHeader().Height
	tmTime := request.GetHeader().Time
	tmHash := fmt.Sprintf("%x", request.GetHash())

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

	app.state.AppendRoundStats(logger, request)

	// reset valset changes
	app.ValUpdates = make([]abci.ValidatorUpdate, 0)
	height := uint64(tmHeight)
	app.SetHeight(height)

	// Tell the search we have a new block on the way.
	if app.indexer != nil {
		start := time.Now()
		err = errors.Wrap(app.indexer.BeginBlock(request, response, app.GetState()), "indexing")
		duration := time.Since(start)
		logger = logger.WithFields(log.Fields{
			"index.elapsed.ns": duration.Nanoseconds(),
		})
	}

	return
}

// DeliverTx services DeliverTx requests
func (app *App) DeliverTx(request abci.RequestDeliverTx) (response abci.ResponseDeliverTx) {
	var tx metatx.Transactable
	var err error
	var logger log.FieldLogger

	logger = app.requestLogger("DeliverTx", true, logger)

	defer func() {
		logger = logger.WithField("returnCode", code.ReturnCode(response.Code).String())
		if err != nil {
			logger = logger.WithError(err)
			logger.Error("DeliverTx erred")
			response.Log = err.Error()
		} else {
			logger.Info("DeliverTx completed successfully")
		}

		// no matter if they got applied or not, we don't want to persist any thunks
		// past this tx
		app.deferredThunks = nil
	}()

	tx, response.Code, logger, err = app.validateTransactable(request.Tx)
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
		if app.indexer != nil {
			start := time.Now()
			err = errors.Wrap(app.indexer.DeliverTx(request, response, tx, app.GetState()), "indexing")
			duration := time.Since(start)
			logger = logger.WithFields(log.Fields{
				"index.elapsed.ns": duration.Nanoseconds(),
			})
		}
	} else {
		logger = logger.WithField("err.context", "applying transaction")
		response.Code = uint32(code.ErrorApplyingTransaction)
		response.Log = err.Error()
	}
	return
}

// EndBlock updates the validator set
func (app *App) EndBlock(request abci.RequestEndBlock) (response abci.ResponseEndBlock) {
	var err error
	logger := app.logRequest("EndBlock", nil)
	defer func() {
		if err != nil {
			logger = logger.WithError(err)
			logger.Error("EndBlock erred")
		} else {
			logger.Info("EndBlock completed successfully")
		}
	}()
	response.ValidatorUpdates = app.ValUpdates
	logger = logger.WithField("validatorUpdates.qty", len(response.ValidatorUpdates))
	if app.indexer != nil {
		start := time.Now()
		err = errors.Wrap(app.indexer.EndBlock(request, response, app.GetState()), "indexing")
		duration := time.Since(start)
		logger = logger.WithFields(log.Fields{
			"index.elapsed.ns": duration.Nanoseconds(),
		})
	}
	return
}

// Commit saves a new version
//
// Panics if InitChain has not been called.
func (app *App) Commit() (response abci.ResponseCommit) {
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
	} else {
		logger = logger.WithField("commit.status", "skipped: no txs pending")
	}
	logger = logger.WithField("abci.sequence", "end")

	response.Data = app.Hash()

	if app.indexer != nil {
		start := time.Now()
		err = errors.Wrap(app.indexer.Commit(response, app.GetState()), "indexing")
		duration := time.Since(start)
		logger = logger.WithFields(log.Fields{
			"index.elapsed.ns": duration.Nanoseconds(),
		})
	}

	return
}
