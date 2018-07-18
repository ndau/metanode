package state

import (
	"github.com/attic-labs/noms/go/datas"
	util "github.com/oneiro-ndev/noms-util"
)

// GetHeight gets the height offset from the state
func (state *Metastate) GetHeight() uint64 {
	return uint64(state.Height)
}

// SetHeight sets the height offset in the state
func (state *Metastate) SetHeight(db datas.Database, h uint64) {
	state.Height = util.Int(int64(h))
}
