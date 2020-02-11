package app

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


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
