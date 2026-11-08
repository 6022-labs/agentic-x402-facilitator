package errors

import "errors"

// ErrLockNotOwned means a lock expired and its token no longer owns it.
var ErrLockNotOwned = errors.New("lock not owned")
