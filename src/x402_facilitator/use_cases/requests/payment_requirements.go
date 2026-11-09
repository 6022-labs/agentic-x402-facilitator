package requests

import "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/models/enums"

const extraAssetTransferMethod = "assetTransferMethod"

// PaymentRequirements is one accepted way to pay; only Extra's assetTransferMethod is read, the domain is configured.
type PaymentRequirements struct {
	Scheme            string         `json:"scheme"`
	Network           string         `json:"network"`
	Amount            string         `json:"amount"`
	Asset             string         `json:"asset"`
	PayTo             string         `json:"payTo"`
	MaxTimeoutSeconds int            `json:"maxTimeoutSeconds"`
	Extra             map[string]any `json:"extra,omitempty"`
}

// AssetTransferMethod is Extra's assetTransferMethod, eip3009 when absent as the spec says; false when unknown.
func (r *PaymentRequirements) AssetTransferMethod() (enums.AssetTransferMethod, bool) {
	raw, found := r.Extra[extraAssetTransferMethod]
	if !found {
		return enums.AssetTransferMethodEip3009, true
	}
	value, isString := raw.(string)
	var method enums.AssetTransferMethod
	if !isString || method.Set(value) != nil {
		return "", false
	}
	return method, true
}
