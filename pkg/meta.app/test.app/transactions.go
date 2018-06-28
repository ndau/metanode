package testapp

import (
	"fmt"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
)

//go:generate msgp -tests=0

// TxIDs is the testapp transactions
var TxIDs = metatx.TxIDMap{
	metatx.TxID(1): &Add{},
}

// Add transactions add an appropriate amount to the state
type Add struct {
	Qty int
}

var _ metatx.Transactable = (*Add)(nil)

// Validate implements Transactable
func (a Add) Validate(interface{}) error {
	if a.Qty < 0 {
		return fmt.Errorf("Adding negatives is silly")
	}
	return nil
}

// Apply implements Transactable
func (a Add) Apply(appI interface{}) error {
	app := appI.(*TestApp)
	return app.UpdateCount(func(c *uint64) error {
		*c += uint64(a.Qty)
		return nil
	})
}
