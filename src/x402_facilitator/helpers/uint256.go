package helpers

import (
	"math/big"
	"strings"
)

// ParseUint256 accepts a non-negative decimal integer that fits a uint256.
func ParseUint256(value string) (*big.Int, bool) {
	parsed, ok := new(big.Int).SetString(strings.TrimSpace(value), 10)
	if !ok || parsed.Sign() < 0 || parsed.BitLen() > 256 {
		return nil, false
	}
	return parsed, true
}
