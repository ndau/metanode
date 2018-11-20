package testapp

import (
	"encoding/binary"

	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
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
