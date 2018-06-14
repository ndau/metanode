package testapp

import (
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"

	metast "github.com/oneiro-ndev/metanode/pkg/meta.app/meta.state"
	util "github.com/oneiro-ndev/noms-util"
)

// TestState is a super simple test state
type TestState struct {
	Number util.Int
}

var _ metast.State = (*TestState)(nil)

// MarshalNoms implements metast.State
func (t TestState) MarshalNoms(vrw nt.ValueReadWriter) (nt.Value, error) {
	return marshal.Marshal(vrw, t)
}

// UnmarshalNoms implements metast.State
func (t *TestState) UnmarshalNoms(v nt.Value) error {
	return marshal.Unmarshal(v, t)
}
