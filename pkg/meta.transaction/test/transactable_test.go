package tests

import (
	"testing"

	tx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	"github.com/stretchr/testify/require"
)

func TestTransactableToBytes(t *testing.T) {
	sy := Stringy{S: "foo bar bat"}
	_, err := tx.Marshal(&sy, Tmap)
	require.NoError(t, err)
}

func TestTransactableRoundtrip(t *testing.T) {
	sy := Stringy{S: "foo bar bat"}
	iy := Inty{I: 12345}
	sb, err := tx.Marshal(&sy, Tmap)
	require.NoError(t, err)
	ib, err := tx.Marshal(&iy, Tmap)
	require.NoError(t, err)

	require.NotEqual(t, sb, ib)

	sz, err := tx.Unmarshal(sb, Tmap)
	require.NoError(t, err)
	iz, err := tx.Unmarshal(ib, Tmap)
	require.NoError(t, err)

	// sx and iz both implement the Transactable interface,
	// but they should have different concrete types.
	// For testing purposes, stringys are always valid,
	// and intys never are. Let's see if they deserialized
	// properly.
	require.NoError(t, sz.Validate(nil))
	require.Error(t, iz.Validate(nil))
}
