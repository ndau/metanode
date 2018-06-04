// This file contains info/query connection methods for the App

package app

import (
	"github.com/tendermint/abci/types"
)

// Info services Info requests
func (app *App) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	app.logRequest("Info")
	return types.ResponseInfo{
		LastBlockHeight:  int64(app.Height()),
		LastBlockAppHash: app.Hash(),
	}
}

// SetOption sets application options, but is entirely undocumented
func (app *App) SetOption(request types.RequestSetOption) (response types.ResponseSetOption) {
	logger := app.logRequest("SetOption")
	logger.Info("params", "key", request.GetKey(), "value", request.GetValue())
	return
}

// Query determines the current value for a given key
func (app *App) Query(request types.RequestQuery) (response types.ResponseQuery) {
	app.logRequest("Info")
	response.Height = int64(app.Height())
	return
}
