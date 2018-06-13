package state

import (
	"github.com/attic-labs/noms/go/datas"
	nt "github.com/attic-labs/noms/go/types"

	util "github.com/oneiro-ndev/noms-util"
)

const heightOffsetKey = "heightOffset"

// GetHeightOffset gets the height offset from the state
func (state *Metastate) GetHeightOffset() (uint64, error) {
	i64, err := util.IntFromBlob(nt.Struct(*state).Get(heightOffsetKey).(nt.Blob))
	return uint64(i64), err
}

// SetHeightOffset sets the height offset in the state
func (state *Metastate) SetHeightOffset(db datas.Database, ho uint64) {
	blob := util.Int(int64(ho)).ToBlob(db)
	*state = Metastate(nt.Struct(*state).Set(heightOffsetKey, blob))
}
