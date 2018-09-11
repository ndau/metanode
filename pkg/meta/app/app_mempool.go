// This file contains mempool connection methods for the App

package app

import (
	"fmt"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/types"
)

func (app *App) validateTransactable(bytes []byte) (metatx.Transactable, uint32, log.FieldLogger, error) {
	tx, err := metatx.Unmarshal(bytes, app.txIDs)
	rc := uint32(code.OK)
	if err != nil {
		logger := app.logger.WithError(err).WithField("tx bytes", fmt.Sprintf("%x", bytes))
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
func (app *App) CheckTx(bytes []byte) (response types.ResponseCheckTx) {
	_, rc, logger, err := app.validateTransactable(bytes)
	app.logRequest("CheckTx", logger)
	response.Code = rc
	if err != nil {
		response.Log = err.Error()
	}
	return
}
