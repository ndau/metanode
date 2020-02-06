// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


package app

import metast "github.com/oneiro-ndev/metanode/pkg/meta/state"

// A Thunk is a unit of computation prepared in one place and executed in
// another. It turns out that we need to be able to register and playback
// thunks, because it is important all of a transaction's state updates
// are applied atomically: if any of them fail, none are applied. What we'd
// ideally want is to change definition of `metatx.Apply` from
//
//   Apply(app interface{}) error
//
// to
//
//   Apply(app interface{}) ([]func(metast.State) (metast.State, error), error)
//
// while simultaneously deprecating `UpdateState`, eliminating it from
// node code. The result of this change would be that we could ensure that
// all state updates happened automatically and atomically: each tx application
// would simply return a list of thunks, and the state changes from each
// would be persisted if all thunks returned without error.
//
// We can't do that: our blockchain's history includes some transactions
// for which a previously buggy implementation leaked state deltas despite
// the tx overall failing. Changing the interface in that way would preclude
// our feature-gated workaround, which imposes correct behavior in the future
// while ensuring that playback proceeds properly for those older blocks.
//
// Instead, we have to settle for a somewhat worse solution. It accomplishes
// more or less the same thing, just in a more complicated, somewhat harder-
// to-use way. We can register deferred thunks. Each tx may still call
// UpdateState exactly once, but it may also register an arbitrary number of
// thunks. If the tx.Apply call succeeds, these thunks will be played back
// sequentially, further updating the application state.
//
// However, these deferred thunks have more stringent restrictions than
// normal state updaters. They may not error, and they may not return nil.
// They have to succeed in all cases.
type Thunk func(metast.State) metast.State

// Defer state-modification thunk(s) for application only if the rest of the tx's
// application succeeds.
func (app *App) Defer(thunks ...Thunk) {
	app.deferredThunks = append(app.deferredThunks, thunks...)
}
