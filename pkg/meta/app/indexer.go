package app

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	abci "github.com/tendermint/tendermint/abci/types"
)

// An Indexer is a secondary database which can index blockchain data.
//
// It probably has extra methods with which to search the database, but we
// don't care about those here.
type Indexer interface {
	InitChain(abci.RequestInitChain, abci.ResponseInitChain, metast.State)
	BeginBlock(abci.RequestBeginBlock, abci.ResponseBeginBlock, metast.State)
	DeliverTx(abci.RequestDeliverTx, abci.ResponseDeliverTx, metatx.Transactable, metast.State) // only called for valid txs
	EndBlock(abci.RequestEndBlock, abci.ResponseEndBlock, metast.State)
	Commit(abci.ResponseCommit, metast.State)
}
