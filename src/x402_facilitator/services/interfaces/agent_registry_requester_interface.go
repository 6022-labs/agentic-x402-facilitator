package interfaces

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

// AgentRegistryRequesterInterface reads the 6022 agent registry of every chain the facilitator knows.
type AgentRegistryRequesterInterface interface {
	// IsAgentAddress reports whether address is a registered EVM address of a 6022 agent on any of those chains;
	// it fails only when none matched and one could not be read.
	IsAgentAddress(ctx context.Context, address common.Address) (bool, error)
	// RegistryChainIds lists, sorted, the chains whose registries IsAgentAddress reads.
	RegistryChainIds() []uint64
}
