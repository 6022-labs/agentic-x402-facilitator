package errors

import "errors"

// ErrRelayerUnavailable means the relayer key could not be loaded, so nothing can be simulated or sent.
var ErrRelayerUnavailable = errors.New("relayer unavailable")
