package app

import "github.com/pkg/errors"

func (a *App) invalidChildStateError() error {
	return errors.Wrap(a.childStateValidity, "application state currently invalid, try again later")
}

// SetStateValidity will prevent further action by this node until the state
// error has cleared if a non-nil error is submitted.
//
// Specifically, it will refuse to answer any queries, or validate any transactions.
func (a *App) SetStateValidity(err error) {
	a.childStateValidity = err
}
