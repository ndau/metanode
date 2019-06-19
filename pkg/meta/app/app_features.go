package app

// Feature allows us to define features by numeric value.
type Feature int

// Features allows us to gate feature logic based on app state.
//
// For example, if we change validation rules for a transaction, we have to remain backward
// compatible for playback purposes, in case there are transactions already on the blockchain
// that got there using the outdated validation rules.  Once we upgrade nodes with the new
// feature logic, we can do so once we're past a given block height.  IsActive() can return true
// once that happens.
type Features interface {
	// IsActive returns whether the given feature is currently active.
	//
	// Once a feature becomes "active", it can never become "inactive".  We can handle this when
	// we add more features that override previous features by checking the newest features first.
	//
	// For example, say we add a feature in some transaction validation code that rounded a qty:
	//
	//   qty := math.Round(tx.Qty)
	//
	// Then later we decided to round to the nearest tenth instead, we would write:
	//
	//   qty := tx.Qty
	//   if features.IsActive("RoundToTenths") {
	//       qty = math.Round(qty*10)/10
	//   } else {
	//       qty = math.Round(qty)
	//   }
	//
	// Then even later we decide to round to the nearest hundredth, we would write:
	//
	//   qty := tx.Qty
	//   if features.IsActive("RoundToHundredths") {
	//       qty = math.Round(qty*100)/100
	//   } else if features.IsActive("RoundToTenths") {
	//       qty = math.Round(qty*10)/10
	//   } else {
	//       qty = math.Round(qty)
	//   }
	//
	// That way we remain backward compatible until the new rules become active as the app's
	// state (e.g. block height) increases.
	//
	//   height:        0          120               300
	//                  |           |                 |
	//   blockchain:    |--x---x----+---y------y------+--z--z-------z---...
	//                  |           |                 |
	//   feature:    genesis   RoundToTenths   RoundToHundredths
	//
	// A transaction "x" that occurs prior to block 120 gets the default handling since genesis.
	// A transaction "y" with height in [120, 300) gets the rounding-by-tenths handling.
	// A transaction "z" on or after block height 300 gets the rounding-by-hundredths handling.
	IsActive(feature Feature) bool
}
