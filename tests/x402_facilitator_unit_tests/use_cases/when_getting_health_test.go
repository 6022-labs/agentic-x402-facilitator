package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/use_cases"
	"github.com/stretchr/testify/assert"
)

func TestWhenGettingHealth(t *testing.T) {
	t.Parallel()

	t.Run("Given the process is serving", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report ok", func(t *testing.T) {
			t.Parallel()

			response, err := use_cases.NewGetHealth().Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "ok", response.Status)
		})
	})
}
