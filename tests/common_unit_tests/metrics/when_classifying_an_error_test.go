package metrics_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/6022-labs/agentic-x402-facilitator/src/common/metrics"
	"github.com/stretchr/testify/assert"
)

func TestWhenClassifyingAnError(t *testing.T) {
	t.Parallel()

	t.Run("Given no error", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify as empty so the attribute is omitted", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "", metrics.ErrorType(nil))
		})
	})

	t.Run("Given a timeout", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify a deadline, a wrapped deadline and an os deadline alike", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "timeout", metrics.ErrorType(context.DeadlineExceeded))
			assert.Equal(t, "timeout", metrics.ErrorType(fmt.Errorf("dial: %w", context.DeadlineExceeded)))
			assert.Equal(t, "timeout", metrics.ErrorType(os.ErrDeadlineExceeded))
		})
	})

	t.Run("Given a cancellation", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify as canceled", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "canceled", metrics.ErrorType(context.Canceled))
		})
	})

	t.Run("Given a name resolution failure", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify as dns", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "dns", metrics.ErrorType(&net.DNSError{Err: "no such host"}))
		})
	})

	t.Run("Given a connection failure", func(t *testing.T) {
		t.Parallel()

		t.Run("Should classify as connection", func(t *testing.T) {
			t.Parallel()

			opErr := &net.OpError{Op: "dial", Err: errors.New("connection refused")}
			assert.Equal(t, "connection", metrics.ErrorType(opErr))
		})
	})

	t.Run("Given an error with no lower-cardinality classification", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fall back to the semconv catch-all", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, metrics.ErrorTypeOther, metrics.ErrorType(errors.New("boom")))
		})
	})
}
