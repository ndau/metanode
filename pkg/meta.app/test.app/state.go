package testapp

import (
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	metast "github.com/oneiro-ndev/metanode/pkg/meta.app/meta.state"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
)

// TestState is a super simple test state
type TestState struct {
	Number util.Int
}

var _ metast.State = (*TestState)(nil)

// MarshalNoms implements metast.State
func (t TestState) MarshalNoms(vrw nt.ValueReadWriter) (nt.Value, error) {
	numValue, err := t.Number.MarshalNoms(vrw)
	if err != nil {
		return nil, err
	}
	return marshal.Marshal(vrw, nt.NewStruct("TestState", nt.StructData{
		"Number": numValue,
	}))
}

// UnmarshalNoms implements metast.State
func (t *TestState) UnmarshalNoms(v nt.Value) (err error) {
	strct, isStruct := v.(nt.Struct)
	if !isStruct {
		return errors.New("TestState.UnmarshalNoms: v is not a struct")
	}
	numVal, hasNumVal := strct.MaybeGet("Number")
	if !hasNumVal {
		return errors.New("TestState.UnmarshalNoms: Number not found")
	}
	return errors.Wrap(t.Number.UnmarshalNoms(numVal), "TestState.UnmarshalNoms")
}

// Init satisfies metast.State
func (*TestState) Init(vrw nt.ValueReadWriter) {}
