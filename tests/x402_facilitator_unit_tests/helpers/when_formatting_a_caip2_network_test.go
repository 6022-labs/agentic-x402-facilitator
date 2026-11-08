package helpers_test

import (
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/stretchr/testify/assert"
)

func TestWhenFormattingACaip2Network(t *testing.T) {
	t.Parallel()

	t.Run("Given a chain id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should prefix it with the eip155 namespace", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "eip155:137", helpers.FormatCaip2(137))
		})
	})
}
