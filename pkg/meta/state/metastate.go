// Package state contains the metastate, which is the actual state stored by a metaapp.
//
// It stores such data as the validator set and the height offset,
// and wraps the application state.
package state

import (
	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/marshal"
	"github.com/pkg/errors"
)

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// Metastate wraps the client app state and keeps track of bookkeeping data
// such as the validator set and height offset.
//nomsify Metastate
type Metastate struct {
	Validators map[string]int64
	Height     uint64
	Stats      VoteStats
	ChildState State
}

const metastateName = "metastate"

func newMetaState(db datas.Database, child State) Metastate {
	if child == nil {
		panic("programming error: nil child in newMetaState")
	}
	return Metastate{
		ChildState: child,
	}
}

// Load the metastate from a DB and DS
//
// Initialize the DB if it hasn't already been done.
// We need an example of the child state so we can properly initialize everything
func (state *Metastate) Load(db datas.Database, ds datas.Dataset, child State) (datas.Dataset, error) {
	var err error
	if state == nil {
		return ds, errors.New("Metastate.Load() on nil pointer")
	}
	if child == nil {
		return ds, errors.New("Metastate.Load(): child state must not be nil")
	}
	head, hasHead := ds.MaybeHeadValue()
	if !hasHead {
		head, err = marshal.Marshal(db, newMetaState(db, child))
		if err != nil {
			return ds, errors.Wrap(err, "Load failed to marshal metastate")
		}

		// commit the empty head so when we go to get things later, we don't
		// panic due to an empty dataset
		ds, err = db.CommitValue(ds, head)
		if err != nil {
			return ds, errors.Wrap(err, "Load failed to commit new head")
		}
	}
	if state.ChildState == nil {
		// ensure we can unmarshal the child state without a nil pointer exception
		state.ChildState = child
	}

	return ds, state.UnmarshalNoms(head)
}

// Commit the current state and return an updated dataset
func (state *Metastate) Commit(db datas.Database, ds datas.Dataset) (datas.Dataset, error) {
	value, err := state.MarshalNoms(db)
	if err != nil {
		return ds, errors.Wrap(err, "Metastate.Commit->state.MarshalNoms")
	}

	return db.CommitValue(ds, value)
}
