package enums

// AssetTransferMethod is how an exact payment moves the token; only EIP-3009 is settled.
type AssetTransferMethod string

const AssetTransferMethodEip3009 AssetTransferMethod = "eip3009"

func (m AssetTransferMethod) String() string {
	return string(m)
}
