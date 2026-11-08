package requests

type ResourceInfo struct {
	Url         string `json:"url"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// Eip3009Authorization mirrors EIP-3009 TransferWithAuthorization; Nonce is 0x-hex 32 bytes, the rest decimal strings.
type Eip3009Authorization struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	ValidAfter  string `json:"validAfter"`
	ValidBefore string `json:"validBefore"`
	Nonce       string `json:"nonce"`
}

type ExactEvmPayload struct {
	Signature     string                `json:"signature"`
	Authorization *Eip3009Authorization `json:"authorization,omitempty"`
}

// PaymentPayload is the signed payment; v2 names the chosen requirements in Accepted, older clients set Scheme/Network.
type PaymentPayload struct {
	X402Version int                  `json:"x402Version"`
	Resource    *ResourceInfo        `json:"resource,omitempty"`
	Accepted    *PaymentRequirements `json:"accepted,omitempty"`
	Scheme      string               `json:"scheme,omitempty"`
	Network     string               `json:"network,omitempty"`
	Payload     ExactEvmPayload      `json:"payload"`
}

// AcceptedScheme is the scheme the payer signed for, from Accepted or the legacy top-level field.
func (p *PaymentPayload) AcceptedScheme() string {
	if p.Accepted != nil {
		return p.Accepted.Scheme
	}
	return p.Scheme
}

// AcceptedNetwork is the network the payer signed for, from Accepted or the legacy top-level field.
func (p *PaymentPayload) AcceptedNetwork() string {
	if p.Accepted != nil {
		return p.Accepted.Network
	}
	return p.Network
}
