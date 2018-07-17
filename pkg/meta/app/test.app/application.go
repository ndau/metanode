package testapp

import (
	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
)

// TestApp is an application built solely for testing the metanode stuff
type TestApp struct {
	*meta.App
}

// NewTestApp constructs a new TestApp
func NewTestApp() (*TestApp, error) {
	dbSpec := "mem"
	name := "TestApp"
	metaapp, err := meta.NewApp(dbSpec, name, &TestState{}, TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create metaapp")
	}

	app := TestApp{
		metaapp,
	}
	app.App.SetChild(&app)
	return &app, nil
}

// GetCount returns the count in this app
func (t *TestApp) GetCount() uint64 {
	return uint64(t.GetState().(*TestState).Number)
}

// UpdateCount allows callers to modify the count
//
// If the supplied function returns a non-nil error, the app count is unchanged
// and the error is propagated.
func (t *TestApp) UpdateCount(ud func(*uint64) error) error {
	return t.UpdateState(func(st metast.State) (metast.State, error) {
		state := st.(*TestState)
		n := uint64(state.Number)
		err := ud(&n)
		if err == nil {
			state.Number = util.Int(n)
		}
		return state, err
	})
}
