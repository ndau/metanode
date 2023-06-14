// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


// This file contains mempool connection methods for the App

package app

import (
	"fmt"

	"github.com/ndau/metanode/pkg/meta/app/code"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	log "github.com/sirupsen/logrus"
	abci "github.com/oneiro-ndev/tendermint.0.32.3/abci/types"
)

func (app *App) validateTransactable(bytes []byte) (metatx.Transactable, uint32, log.FieldLogger, error) {
	if app.childStateValidity != nil {
		return nil, uint32(code.InvalidNodeState), app.logger.WithError(app.childStateValidity), app.invalidChildStateError()
	}

	tx, err := metatx.Unmarshal(bytes, app.txIDs)
	rc := uint32(code.OK)
	if err != nil {
		logger := app.logger.WithError(err).WithField("tx.bytes", fmt.Sprintf("%x", bytes))
		logger.Info("Encoding error")
		return nil, uint32(code.EncodingError), logger, err
	}
	logger := app.DecoratedTxLogger(tx)
	app.checkChild()
	err = tx.Validate(app.childApp)
	if err != nil {
		logger.WithError(err).Info("invalid tx")
		rc = uint32(code.InvalidTransaction)
		return nil, rc, logger, err
	}
	return tx, rc, logger, nil
}

// CheckTx validates a Transaction
func (app *App) CheckTx(request abci.RequestCheckTx) (response abci.ResponseCheckTx) {
	_, rc, logger, err := app.validateTransactable(request.Tx)
	app.logRequest("CheckTx", logger)
	response.Code = rc
	if err != nil {
		response.Log = err.Error()
	}
	return
}
