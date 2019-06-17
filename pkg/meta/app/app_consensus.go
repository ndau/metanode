// This file contains consensus connection methods for the App

package app

import (
	"fmt"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// IncrementalIndexer declares methods for incremental indexing.
type IncrementalIndexer interface {
	OnBeginBlock(height uint64, blockTime time.Time, tmHash string) error
	OnDeliverTx(tx metatx.Transactable) error
	OnCommit(app *App) error
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
	tmHeight := req.GetHeader().Height
	tmTime := req.GetHeader().Time
	tmHash := fmt.Sprintf("%x", req.GetHash())

	app.blockHeight = uint64(tmHeight)

	var err error
	app.blockTime, err = math.TimestampFrom(tmTime)
	if err != nil {
		app.DecoratedLogger().WithFields(log.Fields{
			"tm.height": tmHeight,
			"tm.time":   tmTime,
			"tm.hash":   tmHash,
		}).WithError(err).Error(
			"failed to create ndau timestamp from block time",
		)
		// we panic because without a good block time, we can't recover
		// The only error conditions for this function are when the timestamp
		// is before the epoch, and overflowing our time type. Neither case
		// is likely on a running blockchain.
		panic(err)
	}

	var logger log.FieldLogger
	logger = app.DecoratedLogger().WithFields(log.Fields{
		"tm.height": tmHeight,
		"tm.time":   tmTime,
		"tm.hash":   tmHash,
	})
	logger = app.logRequest("BeginBlock", logger)

	app.state.AppendRoundStats(logger, req)

	// reset valset changes
	app.ValUpdates = make([]abci.ValidatorUpdate, 0)
	height := uint64(tmHeight)
	app.SetHeight(height)

	// Tell the search we have a new block on the way.
	search := app.GetSearch()
	if search != nil {
		err := search.OnBeginBlock(height, tmTime, tmHash)
		if err != nil {
			logger.WithError(err).Error("Failed to begin block for search")
		}
	}

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

		// Update the search with the new transaction.
		search := app.GetSearch()
		if search != nil {
			err = search.OnDeliverTx(tx)
			if err != nil {
				logger.WithError(err).Error("Failed to deliver tx for search")
				response.Code = uint32(code.IndexingError)
				response.Log = err.Error()
			}
		}
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
	logger = app.DecoratedLogger().WithFields(log.Fields{
		"qty transactions in block": app.transactionsPending,
		"intra-ABCI log sequence":   "start",
	})
	app.logRequest("Commit", logger)

	if app.transactionsPending > 0 {
		app.transactionsPending = 0
		err := app.commit(logger)
		if err != nil {
			logger.WithError(err).WithField("intra-ABCI log sequence", "mid").Error("Failed to commit block")
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

		// Index the transactions in the new block.
		search := app.GetSearch()
		if search != nil {
			err = search.OnCommit(app)
			if err != nil {
				logger.WithError(err).WithField("intra-ABCI log sequence", "mid").Error("Failed to commit for search")
			}
		}

		logger.WithField("intra-ABCI log sequence", "mid").Info("Committed noms block")
	} else {
		logger.WithField("intra-ABCI log sequence", "mid").Info("Skipped noms commit")
	}

	logger = app.DecoratedLogger().WithField("intra-ABCI log sequence", "end")
	app.logRequest("Commit", logger)
	return abci.ResponseCommit{Data: app.Hash()}
}
