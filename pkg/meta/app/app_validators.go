package app

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


import (
	"strings"

	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// UpdateValidator updates the app's internal state with the given validator
func (app *App) UpdateValidator(v abci.ValidatorUpdate) {
	app.state.UpdateValidator(app.db, v)

	// we only update the changes array after updating the tree
	app.ValUpdates = append(app.ValUpdates, v)
	vuss := make([]string, 0, len(app.ValUpdates))
	for _, vu := range app.ValUpdates {
		vuss = append(vuss, vu.String())
	}
	app.logger.WithFields(log.Fields{
		"validator.power":  v.GetPower(),
		"validator.PubKey": v.GetPubKey(),
		"app.ValUpdates":   strings.Join(vuss, ", "),
	}).Info("UpdateValidator")
}
