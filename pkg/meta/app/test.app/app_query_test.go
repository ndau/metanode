package testapp

import (
	"encoding/binary"
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func Test_valueQuery(t *testing.T) {
	tests := []struct {
		name  string
		value uint64
	}{
		{"zero", 0},
		{"kilo", 1024},
		{"mega", 1024 * 1024},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := NewTestApp()
			require.NoError(t, err)
			require.NoError(t, app.UpdateCount(func(c *uint64) error {
				*c = tt.value
				return nil
			}))

			resp := app.Query(abci.RequestQuery{Path: ValueEndpoint})
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))

			// check returned value
			buffer := make([]byte, 8)
			binary.BigEndian.PutUint64(buffer, tt.value)
			require.Equal(t, buffer, resp.Value)
		})
	}
}
