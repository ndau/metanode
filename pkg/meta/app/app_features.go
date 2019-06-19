package app

// Features allow us to gate feature logic based on app state.
// For example, if we change validation rules for a transaction, we have to remain backward
// compatible for playback purposes, in case there are transactions already on the blockchain
// that got there using the outdated validation rules.  Once we upgrade nodes with the new
// feature logic, we can do so once we're past a given block height.  IsActive() can return true
// once that happens.
type Features interface {
	IsActive(feature string) bool
}

// ZeroHeightFeatures flags all features as active at height 0.
// i.e. all features are active all the time.
// Useful for unit tests, or on networks where we reset the blockchain (e.g. devnet).
type ZeroHeightFeatures struct {
	Features
}

// IsActive implements Features.
func (f *ZeroHeightFeatures) IsActive(feature string) bool {
	return true
}
