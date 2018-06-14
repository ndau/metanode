// Package state contains the metastate, which is the actual state stored by a metaapp.
//
// It stores such data as the validator set and the height offset,
// and wraps the application state.
package state

import (
	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/pkg/errors"

	util "github.com/oneiro-ndev/noms-util"
)

// Metastate wraps the client app state and keeps track of bookkeeping data
// such as the validator set and height offset.
type Metastate struct {
	Validators   nt.Map
	HeightOffset util.Int
	ChildState   State
}

const metastateName = "metastate"

func newMetaState(db datas.Database, child State) Metastate {
	return Metastate{
		Validators:   nt.NewMap(db),
		HeightOffset: util.Int(1),
		ChildState:   child,
	}
}

// Load the metastate from a DB and DS
func (state *Metastate) Load(db datas.Database, ds datas.Dataset, child State) (datas.Dataset, error) {
	var err error
	if state == nil {
		return ds, errors.New("Metastate.Load() on nil pointer")
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
	err = marshal.Unmarshal(head, state)
	if err != nil {
		return ds, errors.Wrap(err, "Load failed to unmarshal head")
	}
	return ds, err
}

// Commit the current state and return an updated dataset
func (state *Metastate) Commit(db datas.Database, ds datas.Dataset) (datas.Dataset, error) {
	value, err := marshal.Marshal(db, state)
	if err != nil {
		return ds, errors.Wrap(err, "Commit failed to marshal metastate")
	}
	return db.CommitValue(ds, value)
}
