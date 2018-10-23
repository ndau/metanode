package app

import (
	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// UpdateValidator updates the app's internal state with the given validator
func (app *App) UpdateValidator(v abci.ValidatorUpdate) {
	app.state.UpdateValidator(app.db, v)

	// we only update the changes array after updating the tree
	app.ValUpdates = append(app.ValUpdates, v)
	app.logger.WithFields(log.Fields{
		"validator.power":  v.GetPower(),
		"validator.PubKey": v.GetPubKey(),
		"app.ValUpdates":   app.ValUpdates,
	}).Info("UpdateValidator")
}
