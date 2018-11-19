// This file contains info/query connection methods for the App

package app

import (
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Info services Info requests
func (app *App) Info(req abci.RequestInfo) (resInfo abci.ResponseInfo) {
	app.logRequest("Info", nil)
	return abci.ResponseInfo{
		LastBlockHeight:  int64(app.Height()),
		LastBlockAppHash: app.Hash(),
	}
}

// SetOption sets application options, but is entirely undocumented
func (app *App) SetOption(request abci.RequestSetOption) (response abci.ResponseSetOption) {
	var logger log.FieldLogger
	logger = app.GetLogger().WithFields(log.Fields{
		"key":   request.GetKey(),
		"value": request.GetValue(),
	})
	app.logRequest("SetOption", logger)
	return
}

var queryHandlers map[string]func(app interface{}, request abci.RequestQuery, response *abci.ResponseQuery)

func init() {
	queryHandlers = make(map[string]func(app interface{}, request abci.RequestQuery, response *abci.ResponseQuery))
}

// RegisterQueryHandler registers a query handler at a particular endpoint
func RegisterQueryHandler(endpoint string, handler func(app interface{}, request abci.RequestQuery, response *abci.ResponseQuery)) {
	queryHandlers[endpoint] = handler
}

// QueryError is a helper to generate a useful response if an error is not nil
func (app *App) QueryError(err error, response *abci.ResponseQuery, msg string) {
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
func (app *App) Query(request abci.RequestQuery) (response abci.ResponseQuery) {
	var logger log.FieldLogger
	logger = app.GetLogger().WithFields(log.Fields{
		"app.height":   app.Height(),
		"query.path":   request.GetPath(),
		"query.data":   request.GetData(),
		"query.height": request.GetHeight(),
	})
	app.logRequest("Query", logger)
	response.Height = int64(app.Height())

	if app.childStateValidity != nil {
		app.QueryError(
			app.invalidChildStateError(),
			&response, "",
		)
		response.Code = uint32(code.InvalidNodeState)
		return
	}

	querykeys := []string{}
	for k := range queryHandlers {
		querykeys = append(querykeys, k)
	}

	handle, hasHandler := queryHandlers[request.GetPath()]
	if !hasHandler {
		response.Code = uint32(code.QueryError)
		response.Log = "Unknown query path"
		logger.WithField("supportedhandlers", querykeys).WithField("requestedPath", request.GetPath()).Error("unknown query path")
		return
	}
	app.checkChild()
	handle(app.childApp, request, &response)

	return
}
