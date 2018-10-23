package state

import (
	"github.com/attic-labs/noms/go/datas"
	nt "github.com/attic-labs/noms/go/types"
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmconv "github.com/tendermint/tendermint/types"
)

// UpdateValidator updates the app's internal state with the given validator
func (state *Metastate) UpdateValidator(db datas.Database, v abci.ValidatorUpdate) error {
	pk := v.GetPubKey()
	pkB, err := pk.Marshal()
	if err != nil {
		return errors.Wrap(err, "UpdateValidator: marshalling update public key")
	}
	pkS := util.Blob(db, pkB)
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
	state.Validators.Iter(func(key, value nt.Value) (stop bool) {
		// extract pubkey from noms
		var pkB []byte
		pkB, err = util.Unblob(key.(nt.Blob))
		if err != nil {
			err = errors.Wrap(err, "GetValidators: unblob public key")
			return true // stop iteration
		}
		pk := abci.PubKey{}
		err = pk.Unmarshal(pkB)
		if err != nil {
			err = errors.Wrap(err, "GetValidators: unmarshal public key")
			return true
		}

		// transform pubkey to address
		// note that this conversion function is specifically marked as UNSTABLE
		// in the TM source, so we should expect this to break.
		// OTOH, there's apparently no stable way to make this conversion happen,
		// so here we are.
		// https://github.com/tendermint/tendermint/blob/0c9c3292c918617624f6f3fbcd95eceade18bcd5/types/protobuf.go#L170-L171
		var tcpk crypto.PubKey
		tcpk, err = tmconv.PB2TM.PubKey(pk)
		if err != nil {
			err = errors.Wrap(err, "GetValidators: convert tm.abci pk into tm.crypto pk")
			return true
		}
		address := tcpk.Address()

		// extract power from noms
		var vblob nt.Blob
		var ok bool
		vblob, ok = value.(nt.Blob)
		if !ok {
			var vbptr *nt.Blob
			vbptr, ok = value.(*nt.Blob)
			if !ok {
				err = errors.New("GetValidators: power not encoded as blob")
				return true
			}
			vblob = *vbptr
		}
		var power util.Int
		power, err = util.IntFromBlob(vblob)
		if err != nil {
			err = errors.Wrap(err, "GetValidators: IntFromBlob power")
			return true
		}

		// finish
		validators = append(validators, abci.Validator{
			Address: address,
			Power:   int64(power),
		})
		return
	})
	return
}
