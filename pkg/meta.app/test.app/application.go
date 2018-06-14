package testapp

import (
	meta "github.com/oneiro-ndev/metanode/pkg/meta.app"
	"github.com/pkg/errors"
)

// TestApp is an application built solely for testing the metanode stuff
type TestApp struct {
	*meta.App
	count uint64
}

// NewTestApp constructs a new TestApp
func NewTestApp(dbSpec string) (*TestApp, error) {
	name := "TestApp"
	metaapp, err := meta.NewApp(dbSpec, name, new(TestState), TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create metaapp")
	}

	app := TestApp{
		metaapp,
		0,
	}
	app.App.SetChild(&app)
	return &app, nil
}

// GetCount returns the count in this app
func (t *TestApp) GetCount() uint64 {
	return t.count
}

// UpdateCount allows callers to modify the count
//
// If the supplied function returns a non-nil error, the app count is unchanged
// and the error is propagated.
func (t *TestApp) UpdateCount(ud func(*uint64) error) error {
	count := t.count
	err := ud(&count)
	if err == nil {
		t.count = count
	}
	return err
}
