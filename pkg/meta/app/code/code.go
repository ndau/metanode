package code

// Use `go generate` to create the ReturnCode stringer
//go:generate stringer -type=ReturnCode

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

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
