package errors

import "errors"

// ErrTransferRejected means the chain would revert the transfer, so nothing was broadcast.
var ErrTransferRejected = errors.New("transfer rejected by the chain")
