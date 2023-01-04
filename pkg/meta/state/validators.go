package state

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


import (
	"encoding/base64"

	"github.com/ndau/noms/go/datas"
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
		return errors.Wrap(err, "Metastate.UpdateValidator->v.PubKey.Marshal")
	}
	pkS := base64.StdEncoding.EncodeToString(pkB)
	if v.Power == 0 {
		delete(state.Validators, pkS)
	} else {
		state.Validators[pkS] = v.GetPower()
	}
	return nil
}

// GetValidators returns a list of validators this app knows of
func (state *Metastate) GetValidators() (validators []abci.Validator, err error) {
	for pubkeyB64, power := range state.Validators {
		pkB, err := base64.StdEncoding.DecodeString(pubkeyB64)
		if err != nil {
			return nil, errors.Wrap(err, "Metastate.GetValidators->decode pubkey b64")
		}
		pk := abci.PubKey{}
		err = pk.Unmarshal(pkB)
		if err != nil {
			return nil, errors.Wrap(err, "GetValidators: unmarshal public key")
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
			return nil, errors.Wrap(err, "GetValidators: convert tm.abci pk into tm.crypto pk")
		}
		address := tcpk.Address()

		// finish
		validators = append(validators, abci.Validator{
			Address: address,
			Power:   power,
		})
	}
	return
}
