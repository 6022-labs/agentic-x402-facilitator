package transports

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	metrics_interfaces "github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_blockchain/metrics/interfaces"
)

// RpcRoundTripper hides the RPC URL from the client, so no error quotes an API key, and measures each call.
type RpcRoundTripper struct {
	base                 http.RoundTripper
	target               *url.URL
	blockchainRpcMetrics metrics_interfaces.BlockchainRpcMetricsInterface
	chainId              string
}

func NewRpcRoundTripper(
	base http.RoundTripper,
	rpcUrl string,
	blockchainRpcMetrics metrics_interfaces.BlockchainRpcMetricsInterface,
	chainId string,
) (*RpcRoundTripper, error) {
	target, err := url.Parse(rpcUrl)
	if err != nil || target.Host == "" {
		return nil, fmt.Errorf("rpc url of chain %s is not a URL", chainId)
	}
	if base == nil {
		base = http.DefaultTransport
	}
	return &RpcRoundTripper{
		base:                 base,
		target:               target,
		blockchainRpcMetrics: blockchainRpcMetrics,
		chainId:              chainId,
	}, nil
}

// Origin is the URL to dial: go-ethereum errors quote it, and providers put API keys in the path.
func (t *RpcRoundTripper) Origin() string {
	return t.target.Scheme + "://" + t.target.Host
}

func (t *RpcRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	resp, err := t.base.RoundTrip(t.forward(req))

	// JSON-RPC errors return HTTP 200 and are not caught here.
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	t.blockchainRpcMetrics.RecordRequest(req.Context(), t.chainId, req.URL.String(), time.Since(start), statusCode, err)

	return resp, err
}

// forward points a copy of req at the full RPC URL.
func (t *RpcRoundTripper) forward(req *http.Request) *http.Request {
	forwarded := req.Clone(req.Context())
	target := *t.target
	target.User = nil
	forwarded.URL = &target
	// http.Client sets basic auth from the URL it was given, which no longer holds the credentials.
	if t.target.User != nil && forwarded.Header.Get("Authorization") == "" {
		password, _ := t.target.User.Password()
		forwarded.SetBasicAuth(t.target.User.Username(), password)
	}
	return forwarded
}
