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
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/oneiro-ndev/tendermint.0.32.3/abci/types"
	tmconv "github.com/oneiro-ndev/tendermint.0.32.3/types"
)

// ValUpdateToVal is a test helper which converts between these equivalent types
func ValUpdateToVal(t *testing.T, vu abci.ValidatorUpdate) abci.Validator {
	cpk, err := tmconv.PB2TM.PubKey(vu.PubKey)
	require.NoError(t, err)
	return abci.Validator{
		Address: cpk.Address(),
		Power:   vu.Power,
	}
}

// ValUpdatesToVals is a test helper which converts slices between these equivalent types
func ValUpdatesToVals(t *testing.T, vus []abci.ValidatorUpdate) []abci.Validator {
	out := make([]abci.Validator, len(vus))
	for i := 0; i < len(vus); i++ {
		out[i] = ValUpdateToVal(t, vus[i])
	}
	return out
}

// ValidatorsAreEquivalent is a test helper which ensures that two lists of
// validators are in fact equivalent.
func ValidatorsAreEquivalent(t *testing.T, a, b []abci.Validator) {
	require.Equal(t, len(a), len(b))
	// convert a and b into maps of public key to power
	// this discards irrelevant tendermint-specific internal data
	// which we don't care about
	aMap := make(map[string]int64)
	bMap := make(map[string]int64)
	for i := 0; i < len(a); i++ {
		aMap[string(a[i].Address)] = a[i].Power
		bMap[string(b[i].Address)] = b[i].Power
	}
	require.Equal(t, len(a), len(aMap), "a list had duplicate public keys")
	require.Equal(t, len(b), len(bMap), "b list had duplicate public keys")
	require.Equal(t, aMap, bMap)
}
