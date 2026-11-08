package settings

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

//go:embed agent_collections_managers.json
var agentCollectionsManagersJson []byte

type agentCollectionsManagerEntry struct {
	ChainId uint64 `json:"chainId"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type agentCollectionsManagersFile struct {
	Managers []agentCollectionsManagerEntry `json:"managers"`
}

// AgentCollectionsManagersSettings lists the registries baked into the build, not operator config.
type AgentCollectionsManagersSettings map[uint64]common.Address

func NewAgentCollectionsManagersSettings(logger *zap.Logger) *AgentCollectionsManagersSettings {
	var f agentCollectionsManagersFile
	if err := json.Unmarshal(agentCollectionsManagersJson, &f); err != nil {
		logger.Fatal("failed to parse built-in agent_collections_managers.json", zap.Error(err))
	}

	settings := AgentCollectionsManagersSettings{}
	for _, m := range f.Managers {
		addr := strings.TrimSpace(m.Address)
		if addr == "" {
			continue
		}
		if !common.IsHexAddress(addr) {
			logger.Fatal(
				"agent_collections_manager address is not a valid hex address",
				zap.Uint64("chain_id", m.ChainId),
				zap.String("address", addr),
			)
		}
		settings[m.ChainId] = common.HexToAddress(addr)
	}

	return &settings
}

func (s *AgentCollectionsManagersSettings) Get(chainId uint64) (common.Address, bool) {
	if s == nil {
		return common.Address{}, false
	}
	address, ok := (*s)[chainId]
	return address, ok
}
