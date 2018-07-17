package app

import (
	"encoding/hex"

	"github.com/attic-labs/noms/go/hash"
)

// bytesOfHash gets an actual byte slice out of a noms hash.Hash
//
// Why is this a separate function instead of operating inline?
// Ask the go compiler: as its own function, this works; inline,
// it gerates a compiler error:
//
// invalid operation app.ds.HeadRef().valueImpl.Hash()[:] (slice of unaddressable value)
func bytesOfHash(hash hash.Hash) []byte {
	return []byte(hash[:])
}

// Hash returns the current hash of the dataset
func (app *App) Hash() []byte {
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
