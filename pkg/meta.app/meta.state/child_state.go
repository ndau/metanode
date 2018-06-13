package state

import (
	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/pkg/errors"
)

const childStateKey = "childState"

// State is a generic application state, backed by noms.
//
// For the most part, actual application implementations will
// use type assertions to their particular state implementation
// to gain access to the full range of state functions for that
// application.
type State interface {
	marshal.Marshaler
	marshal.Unmarshaler
}

// GetChild updates the argument with the current value from the state
func (state *Metastate) GetChild(child State) error {
	if state == nil {
		return errors.New("GetChild called on nil state")
	}
	if child == nil {
		return errors.New("GetChild called with nil child argument")
	}
	return child.UnmarshalNoms(nt.Struct(*state).Get(childStateKey))
}

// SetChild updates the state with the provided child state
func (state *Metastate) SetChild(db datas.Database, child State) error {
	if state == nil {
		return errors.New("SetChild called on nil state")
	}
	if child == nil {
		return errors.New("SetChild called with nil child argument")
	}
	value, err := child.MarshalNoms(db)
	if err != nil {
		return err
	}
	*state = Metastate(nt.Struct(*state).Set(childStateKey, value))
	return nil
}
