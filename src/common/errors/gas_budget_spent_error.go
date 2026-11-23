package errors

import "errors"

// ErrGasBudgetSpent means the chain's daily gas budget cannot cover the settlement, so nothing was broadcast.
var ErrGasBudgetSpent = errors.New("daily gas budget spent")
