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
	// we need some custom unmarshal logic here
	// child is an interface, and Unmarshal has no idea what concrete type
	// goes with that interface. We, however, possess that information.
	// Let's put it to use.
	nStruct, isNStruct := head.(nt.Struct)
	if !isNStruct {
		return ds, errors.New("Load failed: head is not noms Struct")
	}
	validators, hasValidators := nStruct.MaybeGet("validators")
	if !hasValidators {
		return ds, errors.New("Load failed: noms Struct has no Validators")
	}
	valMap, isValMap := validators.(nt.Map)
	if !isValMap {
		return ds, errors.New("Load failed: noms Validators are not a Map")
	}
	state.Validators = valMap
	heightOffset, hasHeightOffset := nStruct.MaybeGet("heightOffset")
	if !hasHeightOffset {
		return ds, errors.New("Load failed: noms Struct has no HeightOffset")
	}
	err = state.HeightOffset.UnmarshalNoms(heightOffset)
	if err != nil {
		return ds, errors.Wrap(err, "Load failed")
	}
	state.ChildState = child
	childState, hasChildState := nStruct.MaybeGet("childState")
	if !hasChildState {
		return ds, errors.New("Load failed: noms Struct has no ChildState")
	}
	err = state.ChildState.UnmarshalNoms(childState)
	if err != nil {
		return ds, errors.Wrap(err, "Load failed while unmarshaling child")
	}

	return ds, err
}

// Commit the current state and return an updated dataset
func (state *Metastate) Commit(db datas.Database, ds datas.Dataset) (datas.Dataset, error) {
	// marshal.Marshal doesn't work here because we're writing an interface type
	// nothing to it but to do it: let's hand-roll this implementation
	hoValue, err := state.HeightOffset.MarshalNoms(db)
	if err != nil {
		return ds, errors.Wrap(err, "Commit failed to marshal HeightOffset")
	}
	childValue, err := state.ChildState.MarshalNoms(db)
	if err != nil {
		return ds, errors.Wrap(err, "Commit failed to marshal ChildState")
	}

	nStruct := nt.NewStruct(metastateName, nt.StructData{
		"validators":   state.Validators,
		"heightOffset": hoValue,
		"childState":   childValue,
	})

	return db.CommitValue(ds, nStruct)
}
