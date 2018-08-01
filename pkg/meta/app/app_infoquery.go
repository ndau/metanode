// This file contains info/query connection methods for the App

package app

import (
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/types"
)

// Info services Info requests
func (app *App) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	app.logRequest("Info", nil)
	return types.ResponseInfo{
		LastBlockHeight:  int64(app.Height()),
		LastBlockAppHash: app.Hash(),
	}
}

// SetOption sets application options, but is entirely undocumented
func (app *App) SetOption(request types.RequestSetOption) (response types.ResponseSetOption) {
	var logger log.FieldLogger
	logger = app.GetLogger().WithFields(log.Fields{
		"key":   request.GetKey(),
		"value": request.GetValue(),
	})
	app.logRequest("SetOption", logger)
	return
}

var queryHandlers map[string]func(app interface{}, request types.RequestQuery, response *types.ResponseQuery)

func init() {
	queryHandlers = make(map[string]func(app interface{}, request types.RequestQuery, response *types.ResponseQuery))
}

// RegisterQueryHandler registers a query handler at a particular endpoint
func RegisterQueryHandler(endpoint string, handler func(app interface{}, request types.RequestQuery, response *types.ResponseQuery)) {
	queryHandlers[endpoint] = handler
}

// QueryError is a helper to generate a useful response if an error is not nil
func (app *App) QueryError(err error, response *types.ResponseQuery, msg string) {
	if err != nil {
		app.GetLogger().WithError(err).Error(msg)

		if len(msg) > 0 {
			msg = msg + ": "
		}
		msg += err.Error()

		response.Log = msg
		response.Code = uint32(code.QueryError)
	}
}

// Query determines the current value for a given key
func (app *App) Query(request types.RequestQuery) (response types.ResponseQuery) {
	app.logRequest("Info", nil)
	response.Height = int64(app.Height())

	handle, hasHandler := queryHandlers[request.GetPath()]
	if !hasHandler {
		response.Code = uint32(code.QueryError)
		response.Log = "Unknown query path"
		return
	}
	app.checkChild()
	handle(app.childApp, request, &response)

	return
}
