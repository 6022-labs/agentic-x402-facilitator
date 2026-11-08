package helpers

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
)

// executionRevertedCode is the JSON-RPC error code geth-compatible nodes return for a revert carrying data.
const executionRevertedCode = 3

// IsRevert tells a node's "this call reverts" answer apart from a transport or node failure.
func IsRevert(err error) bool {
	if err == nil {
		return false
	}
	var rpcErr rpc.Error
	if errors.As(err, &rpcErr) && rpcErr.ErrorCode() == executionRevertedCode {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "execution reverted")
}
