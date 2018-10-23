// Package state contains the metastate, which is the actual state stored by a metaapp.
//
// It stores such data as the validator set and the height offset,
// and wraps the application state.
package state

import (
	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
)

// Metastate wraps the client app state and keeps track of bookkeeping data
// such as the validator set and height offset.
type Metastate struct {
	Validators nt.Map
	Height     util.Int
	ChildState State
	Search     *SearchClient
}

const metastateName = "metastate"

func newMetaState(db datas.Database, child State) Metastate {
	return Metastate{
		Validators: nt.NewMap(db),
		Height:     util.Int(0),
		ChildState: child,
		Search:     NewSearchClient(),
	}
}

// we need some custom unmarshal logic here
// child is an interface, and Unmarshal has no idea what concrete type
// goes with that interface. We, however, possess that information.
// Let's put it to use.
func (state *Metastate) unmarshal(val nt.Value, child State) error {
	nStruct, isNStruct := val.(nt.Struct)
	if !isNStruct {
		return errors.New("Load failed: val is not noms Struct")
	}
	validators, hasValidators := nStruct.MaybeGet("validators")
	if !hasValidators {
		return errors.New("Load failed: noms Struct has no Validators")
	}
	valMap, isValMap := validators.(nt.Map)
	if !isValMap {
		return errors.New("Load failed: noms Validators are not a Map")
	}
	state.Validators = valMap
	height, hasHeight := nStruct.MaybeGet("height")
	if !hasHeight {
		return errors.New("Load failed: noms Struct has no Height")
	}
	err := state.Height.UnmarshalNoms(height)
	if err != nil {
		return errors.Wrap(err, "Load failed")
	}
	state.ChildState = child
	childState, hasChildState := nStruct.MaybeGet("childState")
	if !hasChildState {
		return errors.New("Load failed: noms Struct has no ChildState")
	}
	err = state.ChildState.UnmarshalNoms(childState)
	if err != nil {
		return errors.Wrap(err, "Load failed while unmarshaling child")
	}

	return err
}

// Load the metastate from a DB and DS
// We need an example of the child state so we can properly unmarshal everything
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

	if state.Search == nil {
		state.Search = NewSearchClient()
	}

	return ds, state.unmarshal(head, child)
}

// Commit the current state and return an updated dataset
func (state *Metastate) Commit(db datas.Database, ds datas.Dataset) (datas.Dataset, error) {
	// marshal.Marshal doesn't work here because we're writing an interface type
	// nothing to it but to do it: let's hand-roll this implementation
	heightValue, err := state.Height.MarshalNoms(db)
	if err != nil {
		return ds, errors.Wrap(err, "Commit failed to marshal Height")
	}
	childValue, err := state.ChildState.MarshalNoms(db)
	if err != nil {
		return ds, errors.Wrap(err, "Commit failed to marshal ChildState")
	}

	nStruct := nt.NewStruct(metastateName, nt.StructData{
		"validators": state.Validators,
		"height":     heightValue,
		"childState": childValue,
	})

	return db.CommitValue(ds, nStruct)
}
