// Package fixtures deploys the contract doubles the blockchain integration tests run against; see README.md.
package fixtures

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator/helpers"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-x402-facilitator/tests/common_tests/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

//go:embed artifacts/*.json
var artifactFiles embed.FS

type artifacts struct {
	files embed.FS
}

type Eip3009Token struct {
	Address  common.Address
	deployer *contracts.Deployer
	contract *bind.BoundContract
}

type AgentCollectionsManager struct {
	Address  common.Address
	deployer *contracts.Deployer
	contract *bind.BoundContract
}

func DeployEip3009Token(
	ctx context.Context,
	client *ethclient.Client,
	anvilAccountSettings container_builders.AnvilAccountSettings,
	eip712Name string,
	eip712Version string,
) (*Eip3009Token, error) {
	deployer, compiled, err := (&artifacts{files: artifactFiles}).prepare(
		ctx,
		client,
		anvilAccountSettings,
		"TestEip3009Token",
	)
	if err != nil {
		return nil, err
	}
	address, contract, err := deployer.Deploy("TestEip3009Token", compiled, eip712Name, eip712Version)
	if err != nil {
		return nil, err
	}
	return &Eip3009Token{Address: address, deployer: deployer, contract: contract}, nil
}

func (t *Eip3009Token) Mint(to common.Address, value *big.Int) error {
	return t.deployer.Transact(t.contract, "mint", to, value)
}

func (t *Eip3009Token) SetAllowance(owner common.Address, spender common.Address, value *big.Int) error {
	return t.deployer.Transact(t.contract, "setAllowance", owner, spender, value)
}

func (t *Eip3009Token) BalanceOf(owner common.Address) (*big.Int, error) {
	balance := new(big.Int)
	err := t.deployer.Call(t.contract, &balance, "balanceOf", owner)
	return balance, err
}

func DeployAgentCollectionsManager(
	ctx context.Context,
	client *ethclient.Client,
	anvilAccountSettings container_builders.AnvilAccountSettings,
) (*AgentCollectionsManager, error) {
	deployer, compiled, err := (&artifacts{files: artifactFiles}).prepare(
		ctx,
		client,
		anvilAccountSettings,
		"TestAgentCollectionsManager",
	)
	if err != nil {
		return nil, err
	}
	address, contract, err := deployer.Deploy("TestAgentCollectionsManager", compiled)
	if err != nil {
		return nil, err
	}
	return &AgentCollectionsManager{Address: address, deployer: deployer, contract: contract}, nil
}

// Register stores the address lowercased, as AgentCollectionV1 canonicalizes EVM values.
func (m *AgentCollectionsManager) Register(address common.Address, tokenId int64) error {
	return m.deployer.Transact(m.contract, "register", strings.ToLower(address.Hex()), big.NewInt(tokenId))
}

// InstallPermit2 puts the Permit2 doubles at the canonical addresses the facilitator hard-codes.
func InstallPermit2(ctx context.Context, client *ethclient.Client) error {
	files := &artifacts{files: artifactFiles}
	for name, address := range map[string]string{
		"TestPermit2":               helpers.Permit2Address,
		"TestX402ExactPermit2Proxy": helpers.X402ExactPermit2ProxyAddress,
	} {
		code, err := files.deployedBytecode(name)
		if err != nil {
			return err
		}
		if err := client.Client().CallContext(ctx, nil, "anvil_setCode", address, code); err != nil {
			return fmt.Errorf("install %s: %w", name, err)
		}
	}
	return nil
}

func (a *artifacts) deployedBytecode(name string) (string, error) {
	raw, err := a.files.ReadFile("artifacts/" + name + ".json")
	if err != nil {
		return "", err
	}
	var artifact struct {
		DeployedBytecode string `json:"deployedBytecode"`
	}
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return "", err
	}
	return artifact.DeployedBytecode, nil
}

func (a *artifacts) prepare(
	ctx context.Context,
	client *ethclient.Client,
	anvilAccountSettings container_builders.AnvilAccountSettings,
	name string,
) (*contracts.Deployer, *contracts.CompiledContract, error) {
	deployer, err := contracts.NewDeployer(ctx, client, anvilAccountSettings)
	if err != nil {
		return nil, nil, err
	}
	raw, err := a.files.ReadFile("artifacts/" + name + ".json")
	if err != nil {
		return nil, nil, fmt.Errorf("no committed artifact for %s (see README.md): %w", name, err)
	}
	compiled, err := contracts.Parse(raw)
	if err != nil {
		return nil, nil, err
	}
	return deployer, compiled, nil
}
