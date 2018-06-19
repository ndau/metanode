package state

import (
	"github.com/attic-labs/noms/go/datas"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"

	util "github.com/oneiro-ndev/noms-util"
)

// UpdateValidator updates the app's internal state with the given validator
func (state *Metastate) UpdateValidator(db datas.Database, v abci.Validator) {
	pkBlob := util.Blob(db, v.GetPubKey())
	if v.Power == 0 {
		state.Validators = state.Validators.Edit().Remove(pkBlob).Map()
	} else {
		powerBlob := util.Int(v.Power).ToBlob(db)
		state.Validators = state.Validators.Edit().Set(pkBlob, powerBlob).Map()
	}
}

// GetValidators returns a list of validators this app knows of
func (state *Metastate) GetValidators() (validators []abci.Validator, err error) {
	state.Validators.IterAll(func(key, value nt.Value) {
		// this iterator interface doesn't allow for early failures,
		// so we need to just skip work if an error occurs
		if err != nil {
			return
		}
		pubKey, err := util.Unblob(key.(nt.Blob))
		err = errors.Wrap(err, "GetValidators found non-`nt.Blob` public key")
		if err != nil {
			return
		}
		power, err := util.IntFromBlob(value.(nt.Blob))
		err = errors.Wrap(err, "GetValidators found non-`nt.Blob` power")
		if err != nil {
			return
		}
		validators = append(validators, abci.Validator{
			PubKey: pubKey,
			Power:  int64(power),
		})
	})
	return
}
