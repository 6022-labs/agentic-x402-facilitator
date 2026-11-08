// Package contracts parses compiled contract artifacts and deploys them onto a throwaway chain.
package contracts

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// CompiledContract is a contract artifact reduced to what a deployment needs.
type CompiledContract struct {
	Abi      abi.ABI
	Bytecode []byte
}

// Parse reads a trimmed artifact (`{contractName, abi, bytecode}`); each test tree embeds its own.
func Parse(raw []byte) (*CompiledContract, error) {
	var artifact struct {
		Abi      json.RawMessage `json:"abi"`
		Bytecode string          `json:"bytecode"`
	}
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return nil, fmt.Errorf("failed to parse contract artifact: %w", err)
	}

	parsedAbi, err := abi.JSON(strings.NewReader(string(artifact.Abi)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract abi: %w", err)
	}

	return &CompiledContract{Abi: parsedAbi, Bytecode: common.FromHex(artifact.Bytecode)}, nil
}
