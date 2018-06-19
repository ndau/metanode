package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/abci/types"
)

// UpdateValidator updates the app's internal state with the given validator
func (app *App) UpdateValidator(v types.Validator) {
	logger := app.logger.WithFields(log.Fields{"method": "updateValidator"})
	logger.Info("entered method", "Power", v.GetPower(), "PubKey", v.GetPubKey())
	app.state.UpdateValidator(app.db, v)

	// we only update the changes array after updating the tree
	app.ValUpdates = append(app.ValUpdates, v)
	logger.Info("exiting OK", "app.ValUpdates", app.ValUpdates)
}
