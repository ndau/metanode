// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// This file contains the basic definition for an ABCI Application.
//
// Interface: https://godoc.org/github.com/tendermint/tendermint/abci/types#Application

package app

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/attic-labs/noms/go/d"
	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/spec"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ensure the app is a TM Application
var _ abci.Application = (*App)(nil)

// App is an ABCI application which implements an Oneiro chain
type App struct {
	abci.BaseApplication

	// We're using noms, which isn't quite like traditional
	// relational databases. In particular, we can't simply
	// store the database, get a cursor, and use the db's stateful
	// nature to keep track of what table we're modifying.
	//
	// Instead, noms breaks things down a bit differently:
	// the database object manages communication with the server,
	// and most history; the dataset is the working set with
	// which we make updates and then push commits.
	//
	// We therefore need to store both.
	db datas.Database
	ds datas.Dataset

	// Why store this state at all? Why not just have an app.state() function
	// which constructs it in realtime from app.ds.HeadValue?
	//
	// We want to ensure that at all times, the 'official' state committed
	// into the dataset is only updated on a 'Commit' transaction. This
	// in turn means that we need to persist the state between transactions
	// in memory, which means keeping track of this state object.
	state metast.Metastate

	// Access to blockchain indexing and searching
	search IncrementalIndexer

	// List of pending validator updates
	ValUpdates []abci.ValidatorUpdate

	// This logger captures various ABCI events
	logger log.FieldLogger

	// the name of this application
	name string

	// a map of txid to example struct
	txIDs metatx.TxIDMap

	// the child application, which defines the transactables and state
	// we need this to pass through into the transactables' methods.
	childApp interface{}

	// the child state cache. this is mainly used to store the state's type
	childState metast.State

	// Noms and Tendermint both count tree height from 0. They must agree
	// at all times. However, there are at least two points at which
	// state managers will typically increment the noms height:
	//
	// - during state.Load (inside NewApp), to ensure the dataset has a head
	// - during InitChain, to commit the initial validator set
	//
	// Neither of those occasions also includes a Tendermint height
	// increase. We need to keep them in sync.
	//
	// Worse, due to a combination of tendermint limitations and efficiency
	// concerns, we don't want to create a noms block if a tendermint block
	// happens to be empty. This means that ultimately, the simplest solution
	// is to just store the tendermint height every time we happen to commit,
	// and cache it otherwise.
	height uint64

	// tendermint ignores the empty block creation and interval settings
	// when the application hash changes, because it determines whether
	// or not a block is empty by the app hash, not by counting the
	// actual transactions.
	//
	// The solution is to only actually commit a noms block when there
	// are transactions pending. This variable keeps track of that.
	transactionsPending uint64

	// sometimes child apps need to signal that their state has become temporarily
	// invalid without dying entirely. The canonical example of this is when
	// a ndau node's SVI update times out: that app can't reliably answer
	// queries or validate or post transactions, but at the same time, it's
	// expected to come back up shortly.
	//
	// The metanode can't know when this happens, or when the child node recovers.
	// It therefore just stores this value and returns appropriate codes to all
	// messages until the state comes back up.
	//
	// This value is an interface, which defaults to nil. This allows code which
	// doesn't need this functionality, such as the chaos node, to simply ignore it.
	childStateValidity error

	// official chain time of the current block
	blockTime math.Timestamp
}

// NewApp prepares a new App
//
// - `dbSpec` is the database spec string; empty or "mem" for in-memory,
//     the connection path (parseable by noms)
// - `name` is the name of this app
// - `childState` is the child state manager. It must be initialized to its zero value.
// - `txIDs` is the map of transaction ids to example structs
func NewApp(dbSpec string, name string, childState metast.State, txIDs metatx.TxIDMap) (*App, error) {
	return NewAppWithLogger(dbSpec, name, childState, txIDs, nil)
}

// NewAppWithLogger prepares a new App
//
// - `dbSpec` is the database spec string; empty or "mem" for in-memory,
//     the connection path (parseable by noms)
// - `name` is the name of this app
// - `childState` is the child state manager. It must be initialized to its zero value.
// - `txIDs` is the map of transaction ids to example structs
func NewAppWithLogger(dbSpec string, name string, childState metast.State, txIDs metatx.TxIDMap, logger log.FieldLogger) (*App, error) {
	if len(dbSpec) == 0 {
		dbSpec = "mem"
	}

	sp, err := spec.ForDatabase(dbSpec)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create noms db")
	}

	var db datas.Database
	// we can fail to connect to noms for a variety of reasons, catch these here and report error
	// we use Try() because noms panics in various places
	err = d.Try(func() {
		db = sp.GetDatabase()
	})
	if err != nil {
		return nil, errors.Wrap(d.Unwrap(err), fmt.Sprintf("NewApp failed to connect to noms db, is noms running at: %s?", dbSpec))
	}

	// initialize the child state
	childState.Init(db)

	// in some ways, a dataset is like a particular table in the db
	ds := db.GetDataset(name)

	state := metast.Metastate{}
	ds, err = state.Load(db, ds, childState)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to load existing state")
	}

	if logger == nil {
		logger = NewLogger()
	}

	now, err := math.TimestampFrom(time.Now())
	if err != nil {
		return nil, errors.Wrap(err, "getting current time as ndau time for initial block time")
	}

	return &App{
		db:        db,
		ds:        ds,
		state:     state,
		search:    nil,
		logger:    logger,
		name:      name,
		txIDs:     txIDs,
		height:    state.Height,
		blockTime: now,
	}, nil
}

// SetHeight updates the app's tendermint height
//
// Under normal circumstances, this should never be called by a child
// application. Tendermint heights are automatically adjusted appropriately
// by the metaapp. This function is public so that test fixtures can be
// constructed with appropriate application heights.
func (app *App) SetHeight(h uint64) {
	app.state.Height = h
	app.height = h
}

// SetChild specifies which child app is using this meta.App.
//
// It is required to be called exactly once, during program initialization.
// It is not part of the normal constructor to ensure that it is possible
// to call NewApp within the constructor of the child.
func (app *App) SetChild(child interface{}) {
	if child == nil {
		panic("nil is invalid in SetChild")
	}
	app.childApp = child
}

func (app *App) checkChild() {
	if app.childApp == nil {
		panic("meta.App.childApp unset. Did you call myApp.App.SetChild()?")
	}
}

// GetDB returns the app's database
func (app *App) GetDB() datas.Database {
	return app.db
}

// GetDS returns the app's dataset
func (app *App) GetDS() datas.Dataset {
	return app.ds
}

// GetName returns the name of the app
func (app *App) GetName() string {
	return app.name
}

// GetState returns the current application state
func (app *App) GetState() metast.State {
	return app.state.ChildState
}

// SetSearch sets the app's incremental indexer
func (app *App) SetSearch(search IncrementalIndexer) {
	app.search = search
}

// GetSearch returns the app's incremental indexer
func (app *App) GetSearch() IncrementalIndexer {
	return app.search
}

// UpdateState updates the current child application state
//
// Returning a nil state from the internal function is an error.
// Returning an error from the internal function returns that error.
func (app *App) UpdateState(updaters ...func(state metast.State) (metast.State, error)) error {
	return app.updateStateInner(false, updaters...)
}

// UpdateStateLeaky updates the child application state whether or not the
// internal updaters fail
//
// This is buggy behavior and this function should NEVER be called in ordinary
// usage. However, we've got a running blockchain for which playback depends on
// replicating past bugs, so we have to be able to choose the old behavior.
func (app *App) UpdateStateLeaky(updaters ...func(state metast.State) (metast.State, error)) error {
	return app.updateStateInner(true, updaters...)
}

func (app *App) updateStateInner(leak bool, updaters ...func(state metast.State) (metast.State, error)) error {
	state := app.GetState()

	if !leak {
		// state is an interface. This means that it is always passed by reference,
		// not by value. This in turn means that any changes an updater makes to
		// the state leaks backwards into the child state, even if an error
		// is returned. This is highly undesirable behavior.
		//
		// The normal recommendation in this case is to manually cast the interface
		// value into its concrete type, so that we can make a copy using normal
		// semantics. Unfortunately, that's impossible in this case: we can't name
		// the concrete type, because it depends on the particular app. Even if
		// we were willing to manually enumerate all blockchains depending on the
		// metaapp, we couldn't name those types, because it would lead to circular
		// import paths.
		//
		// Therefore, we have to use reflection to force go to make a copy.

		// this indirect captures the concrete type of the state object
		indirect := reflect.Indirect(reflect.ValueOf(state))
		// create a new instance of the concrete type
		indirect2 := reflect.New(indirect.Type())
		// set the value of the new indirect to the content of the old
		indirect2.Elem().Set(reflect.ValueOf(indirect.Interface()))
		// set the state variable to the copy
		state = indirect2.Interface().(metast.State)
	}

	for _, updater := range updaters {
		var err error
		state, err = updater(state)
		if err != nil {
			return err
		}
		if state == nil {
			return errors.New("nil state returned from UpdateState")
		}
	}
	app.state.ChildState = state
	return nil
}

// UpdateStateImmediately is like UpdateState, but commits immediately.
//
// It also increments the height offset.
//
// This is useful for inserting mock data etc.
func (app *App) UpdateStateImmediately(updaters ...func(state metast.State) (metast.State, error)) error {
	err := app.UpdateState(updaters...)
	if err != nil {
		return err
	}
	logger := app.GetLogger().WithField("method", "UpdateStateImmediately")
	return app.commit(logger)
}

// Close closes the database connection opened on App creation
func (app *App) Close() error {
	return errors.Wrap(app.db.Close(), "Failed to Close ndau.App")
}

// commit the current application state
//
// This is different from Commit, which processes a Commit Tx!
// However, they're related: think HARD before using this function
// outside of func Commit.
func (app *App) commit(logger log.FieldLogger) (err error) {
	ds, err := app.state.Commit(app.db, app.ds)
	if err == nil {
		app.ds = ds
	}
	logger.WithError(err).Info("meta-application commit")
	return err
}

// Height returns the current height of the application
func (app *App) Height() uint64 {
	return app.height
}

// Validators returns a list of the app's validators.
func (app *App) Validators() ([]abci.Validator, error) {
	return app.state.GetValidators()
}

// Support for closing the app with Ctrl+C when running in a shell.
type sigListener struct {
	sigchan chan os.Signal
}

// We pass in the app, not the logger, in case the logger is changed after we start watching.
func (s *sigListener) watchSignals(app *App) {
	go func() {
		if s.sigchan == nil {
			s.sigchan = make(chan os.Signal, 1)
		}
		signal.Notify(s.sigchan, syscall.SIGTERM, syscall.SIGINT)
		for {
			sig := <-s.sigchan
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				app.GetLogger().Info(fmt.Sprintf("Exiting after receiving '%v' signal", sig))
				os.Exit(0)
			}
		}
	}()
}

// WatchSignals starts a goroutine exits the app gracefully when SIGTERM or SIGINT is received.
func (app *App) WatchSignals() {
	sl := &sigListener{}
	sl.watchSignals(app)
}

// GetStats returns node voting statistics.
//
// This is typically used to update node goodness / voting power.
func (app *App) GetStats() metast.VoteStats {
	return app.state.Stats
}

// BlockTime returns the timestamp of the current block
//
// Note that this can lag fairly significantly behind real time; the only upper
// bound to the lag is the empty block creation rate. As of early 2019, the
// empty block creation rate is 5 minutes, so we expect to see BlockTime up to
// five minutes behind real time. This is due to Tendermint's block model and
// can't be fixed by our code.
func (app *App) BlockTime() math.Timestamp {
	return app.blockTime
}
