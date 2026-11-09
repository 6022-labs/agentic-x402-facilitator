package enums

import (
	"fmt"
	"strings"
)

// AssetTransferMethod is how an exact payment moves the token.
type AssetTransferMethod string

const (
	// AssetTransferMethodEip3009 calls the token's own transferWithAuthorization, signed over its EIP-712 domain.
	AssetTransferMethodEip3009 AssetTransferMethod = "eip3009"
	// AssetTransferMethodPermit2 settles through x402ExactPermit2Proxy, signed over Permit2's fixed domain.
	AssetTransferMethodPermit2 AssetTransferMethod = "permit2"
)

func (m AssetTransferMethod) String() string {
	return string(m)
}

func GetAllAssetTransferMethods() []AssetTransferMethod {
	return []AssetTransferMethod{AssetTransferMethodEip3009, AssetTransferMethodPermit2}
}

func (m *AssetTransferMethod) Set(value string) error {
	normalized := AssetTransferMethod(strings.ToLower(strings.TrimSpace(value)))
	for _, candidate := range GetAllAssetTransferMethods() {
		if normalized == candidate {
			*m = candidate
			return nil
		}
	}
	return fmt.Errorf("invalid AssetTransferMethod: %s", value)
}
