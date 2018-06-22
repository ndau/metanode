// This file contains mempool connection methods for the App

package app

import (
	"fmt"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
	"github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/abci/types"
)

func (app *App) validateTransactable(bytes []byte) (metatx.Transactable, uint32, error) {
	tx, err := metatx.TransactableFromBytes(bytes, app.txIDs)
	rc := uint32(code.OK)
	if err != nil {
		app.logger.WithFields(log.Fields{
			"reason": err.Error(),
			"tx":     fmt.Sprintf("%x", bytes),
		}).Info("Encoding error")
		return nil, uint32(code.EncodingError), err
	}
	app.checkChild()
	err = tx.IsValid(app.childApp)
	if err != nil {
		app.logger.WithField("reason", err.Error()).Info("invalid tx")
		rc = uint32(code.InvalidTransaction)
		return nil, rc, err
	}
	return tx, rc, nil
}

// CheckTx validates a Transaction
func (app *App) CheckTx(bytes []byte) (response types.ResponseCheckTx) {
	app.logger.WithField("type", "CheckTx").Info("Received Request")
	_, rc, err := app.validateTransactable(bytes)
	response.Code = rc
	if err != nil {
		response.Log = err.Error()
	}
	return
}
