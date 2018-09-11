// This file contains the basic definition for an ABCI Application.
//
// Interface: https://godoc.org/github.com/tendermint/tendermint/abci/types#Application

package app

import (
	"fmt"
	"os"

	"github.com/attic-labs/noms/go/d"
	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/spec"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/o11y/pkg/honeycomb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

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

	// List of pending validator updates
	ValUpdates []abci.Validator

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
		logger = log.New()
		logger.(*log.Logger).Formatter = new(log.JSONFormatter)
		logger.(*log.Logger).Out = os.Stderr
		logger = honeycomb.Setup(logger.(*log.Logger))
	}

	return &App{
		db:     db,
		ds:     ds,
		state:  state,
		logger: logger,
		name:   name,
		txIDs:  txIDs,
		height: state.GetHeight(),
	}, nil
}

// SetHeight updates the app's tendermint height
//
// Under normal circumstances, this should never be called by a child
// application. Tendermint heights are automatically adjusted appropriately
// by the metaapp. This function is public so that test fixtures can be
// constructed with appropriate application heights.
func (app *App) SetHeight(h uint64) {
	app.state.SetHeight(app.db, h)
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

// UpdateState updates the current child application state
//
// Returning a nil state from the internal function is an error.
// Returning an error from the internal function returns that error.
func (app *App) UpdateState(updater func(state metast.State) (metast.State, error)) error {
	state, err := updater(app.GetState())
	if err != nil {
		return err
	}
	if state == nil {
		return errors.New("nil state returned from UpdateState")
	}
	app.state.ChildState = state
	return nil
}

// UpdateStateImmediately is like UpdateState, but commits immediately.
//
// It also increments the height offset.
//
// This is useful for inserting mock data etc.
func (app *App) UpdateStateImmediately(updater func(state metast.State) (metast.State, error)) error {
	err := app.UpdateState(updater)
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
