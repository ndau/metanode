// Package state contains the metastate, which is the actual state stored by a metaapp.
//
// It stores such data as the validator set and the height offset,
// and wraps the application state.
package state

import (
	"github.com/attic-labs/noms/go/datas"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/pkg/errors"

	util "github.com/oneiro-ndev/noms-util"
)

// Metastate wraps the client app state and keeps track of bookkeeping data
// such as the validator set and height offset.
type Metastate nt.Struct

const metastateName = "metastate"

// UpdateValidator updates the state's record of the given validator.
//
// When the supplied validator's power is 0, it should be removed.
//	UpdateValidator(datas.Database, abci.Validator)

func newState(db datas.Database, child State) (nt.Struct, error) {
	childValue, err := child.MarshalNoms(db)
	if err != nil {
		return nt.NewStruct("", map[string]nt.Value{}), err
	}
	return nt.NewStruct(metastateName, map[string]nt.Value{
		// Validators is a map of public key to power
		validatorsKey: nt.NewMap(db),
		// heightOffset is the current offset between noms and tendermint height
		heightOffsetKey: util.Int(1).ToBlob(db),
		// childState is an empty child state
		childStateKey: childValue,
	}), nil
}

// Load the metastate from a DB and DS
func (state *Metastate) Load(db datas.Database, ds datas.Dataset, child State) (datas.Dataset, error) {
	var err error
	if state == nil {
		return ds, errors.New("Metastate.Load() on nil pointer")
	}
	head, hasHead := ds.MaybeHeadValue()
	if !hasHead {
		head, err = newState(db, child)
		if err != nil {
			return ds, errors.Wrap(err, "LoadState failed to noms-marshal child")
		}
		// commit the empty head so when we go to get things later, we don't
		// panic due to an empty dataset
		ds, err = db.CommitValue(ds, head)
		if err != nil {
			return ds, errors.Wrap(err, "LoadState failed to commit new head")
		}
	}
	nsS, isS := head.(nt.Struct)
	if !isS {
		return ds, errors.New("LoadState found non-struct as ds.HeadValue")
	}
	*state = Metastate(nsS)
	return ds, err
}

// Commit the current state and return an updated dataset
func (state *Metastate) Commit(db datas.Database, ds datas.Dataset) (datas.Dataset, error) {
	return db.CommitValue(ds, nt.Struct(*state))
}
