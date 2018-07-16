package state

import (
	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// UpdateValidator updates the app's internal state with the given validator
func (state *Metastate) UpdateValidator(db datas.Database, v abci.Validator) error {
	pkS, err := marshal.Marshal(db, v.GetPubKey())
	if err != nil {
		return err
	}
	if v.Power == 0 {
		state.Validators = state.Validators.Edit().Remove(pkS).Map()
	} else {
		powerBlob := util.Int(v.Power).ToBlob(db)
		state.Validators = state.Validators.Edit().Set(pkS, powerBlob).Map()
	}
	return nil
}

// GetValidators returns a list of validators this app knows of
func (state *Metastate) GetValidators() (validators []abci.Validator, err error) {
	state.Validators.IterAll(func(key, value nt.Value) {
		// this iterator interface doesn't allow for early failures,
		// so we need to just skip work if an error occurs
		if err != nil {
			return
		}
		pubKey := abci.PubKey{}
		err := marshal.Unmarshal(key, &pubKey)
		err = errors.Wrap(err, "GetValidators unmarshal public key")
		if err != nil {
			return
		}
		power, err := util.IntFromBlob(value.(nt.Blob))
		err = errors.Wrap(err, "GetValidators IntFromBlob power")
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
