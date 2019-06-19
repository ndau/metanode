package app

// Features allows us to gate feature logic based on app state.
// For example, if we change validation rules for a transaction, we have to remain backward
// compatible for playback purposes, in case there are transactions already on the blockchain
// that got there using the outdated validation rules.  Once we upgrade nodes with the new
// feature logic, we can do so once we're past a given block height.  IsActive() can return true
// once that happens.
type Features interface {
	IsActive(feature string) bool
}
