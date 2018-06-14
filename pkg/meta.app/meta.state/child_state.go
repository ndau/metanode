package state

import (
	"github.com/attic-labs/noms/go/marshal"
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
}
