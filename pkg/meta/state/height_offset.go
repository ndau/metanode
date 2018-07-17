package state

import (
	"github.com/attic-labs/noms/go/datas"
	util "github.com/oneiro-ndev/noms-util"
)

// GetHeightOffset gets the height offset from the state
func (state *Metastate) GetHeightOffset() uint64 {
	return uint64(state.HeightOffset)
}

// SetHeightOffset sets the height offset in the state
func (state *Metastate) SetHeightOffset(db datas.Database, ho uint64) {
	state.HeightOffset = util.Int(int64(ho))
}
