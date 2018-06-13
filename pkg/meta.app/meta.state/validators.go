package state

import (
	"github.com/attic-labs/noms/go/datas"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/pkg/errors"
	abci "github.com/tendermint/abci/types"

	util "github.com/oneiro-ndev/noms-util"
)

const validatorsKey = "validators"

// Validators are the map of public keys to powers in the validator set
func (state *Metastate) validators() nt.Map {
	return nt.Struct(*state).Get(validatorsKey).(nt.Map)
}

// UpdateValidator updates the app's internal state with the given validator
func (state *Metastate) UpdateValidator(db datas.Database, v abci.Validator) {
	validators := state.validators()
	pkBlob := util.Blob(db, v.GetPubKey())
	if v.Power == 0 {
		validators = validators.Edit().Remove(pkBlob).Map()
	} else {
		powerBlob := util.Int(v.Power).ToBlob(db)
		validators = validators.Edit().Set(pkBlob, powerBlob).Map()
	}
	*state = Metastate(nt.Struct(*state).Set(validatorsKey, validators))
}

// GetValidators returns a list of validators this app knows of
func (state *Metastate) GetValidators() (validators []abci.Validator, err error) {
	state.validators().IterAll(func(key, value nt.Value) {
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
