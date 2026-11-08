package metrics

import (
	"context"
	"errors"
	"net"
	"os"
)

// ErrorTypeOther is the semconv catch-all, used only when nothing more specific applies.
const ErrorTypeOther = "_OTHER"

// ErrorType maps an error to a low-cardinality semconv error.type value.
func ErrorType(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, os.ErrDeadlineExceeded):
		return "timeout"
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "dns"
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return "connection"
	}

	return ErrorTypeOther
}
