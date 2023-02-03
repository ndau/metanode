package app

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"encoding/hex"

	"github.com/ndau/noms/go/hash"
)

// bytesOfHash gets an actual byte slice out of a noms hash.Hash
//
// Why is this a separate function instead of operating inline?
// Ask the go compiler: as its own function, this works; inline,
// it gerates a compiler error:
//
// invalid operation app.ds.HeadRef().valueImpl.Hash()[:] (slice of unaddressable value)
func bytesOfHash(hash hash.Hash) []byte {
	// Note - tendermint 0.33 upgrade: Legacy data before genesis
	// On genesis, the noms is preset with date before genesis in testnet and mainnet. This
	// makes it impossible to produce the matched app hash if the new tendermint node wants to
	// sync all of its block from block zero. So this ugly hard coding fix is the work-around
	// solution to allow block-sync to be continued
	h := []byte(hash[:])
	// if hex.EncodeToString(h) == "17af29bb71e0b05c65aa2492337da21d62e9f4d7" {
	// 	fmt.Printf("Pre-genesis data work-around by forcing app hash\r\n")
	// 	h, _ = hex.DecodeString("b5826c2ef692f367051d7a074895ee74f7852afc")
	// }
	// End Note

	return h
}

// Hash returns the current hash of the dataset
func (app *App) Hash() []byte {
	// Note - tendermint 0.33 upgrade:
	// On genesis, the noms has no any dataset. So the app_hash is set to be just an empty string. And
	// the app_hash in the tendermint genesis.json should also be set to empty.
	// On the other hand, when a node starts from a certain block height, noms database must exists already. And
	// so the app_hash would be returned from calculation from noms dataset
	if _, ok := app.ds.MaybeHeadValue(); !ok {
		return []byte("")
	}
	// End Note

	return bytesOfHash(app.ds.HeadRef().Hash())
}

// HashStr returns the current hash of the dataset,
// encoded as a hexadecimal string.
//
// This is useful because Tendermint expects hexadecimal encoding
// for hash strings, but noms by default uses its own sui-generis
// format: the first 20 bytes of big-endian base32 encoding.
func (app *App) HashStr() string {
	return hex.EncodeToString(app.Hash())
}
