package code

// Use `go generate` to create the ReturnCode stringer
//go:generate stringer -type=ReturnCode

// ReturnCode is the type returned by various operations
type ReturnCode uint32

// Return codes for the ndau blockchain
const (
	OK ReturnCode = iota
	InvalidTransaction
	ErrorApplyingTransaction
	EncodingError
	QueryError
	IndexingError
	InvalidNodeState
)
