package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/abci/types"
)

// UpdateValidator updates the app's internal state with the given validator
func (app *App) UpdateValidator(v types.Validator) {
	logger := app.logger.WithFields(log.Fields{
		"method": "updateValidator",
		"power":  v.GetPower(),
		"PubKey": v.GetPubKey(),
	})
	logger.Info("entered method")
	app.state.UpdateValidator(app.db, v)

	// we only update the changes array after updating the tree
	app.ValUpdates = append(app.ValUpdates, v)
	logger.WithField("app.ValUpdates", app.ValUpdates).Info("exiting ok")
}
