package testapp

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


import (
	"encoding/binary"

	meta "github.com/ndau/metanode/pkg/meta/app"
	abci "github.com/tendermint/tendermint/abci/types"
)

func init() {
	meta.RegisterQueryHandler(ValueEndpoint, valueQuery)
}

const ValueEndpoint = "/value"

func valueQuery(appI interface{}, request abci.RequestQuery, response *abci.ResponseQuery) {
	app := appI.(*TestApp)
	value := app.GetCount()
	response.Value = make([]byte, 8)
	binary.BigEndian.PutUint64(response.Value, value)
}
