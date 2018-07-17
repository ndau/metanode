package state

import (
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
)

// State is a generic application state, backed by noms.
//
// For the most part, actual application implementations will
// use type assertions to their particular state implementation
// to gain access to the full range of state functions for that
// application.
type State interface {
	marshal.Marshaler
	marshal.Unmarshaler

	// Init initializes the state.
	//
	// It is expected to be implemented on a pointer receiver, and
	// initialize maps etc as required.
	Init(vrw nt.ValueReadWriter)
}
