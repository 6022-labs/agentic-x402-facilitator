// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abi

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// AgentCollectionInfo is an auto generated low-level Go binding around an user-defined struct.
type AgentCollectionInfo struct {
	Name              string
	Description       string
	ModeratorCount    *big.Int
	NextTokenId       *big.Int
	CollectionAddress common.Address
}

// AgentCollectionsManagerMetaData contains all meta data concerning the AgentCollectionsManager contract.
var AgentCollectionsManagerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"AccessControlBadConfirmation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"neededRole\",\"type\":\"bytes32\"}],\"name\":\"AccessControlUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"collectionId\",\"type\":\"uint256\"}],\"name\":\"CollectionDoesNotExist\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"name\":\"CollectionNameAlreadyExists\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"CollectionSymbolAlreadyExists\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"collectionId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"collectionAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"}],\"name\":\"CollectionCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"CREATOR_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newCollection\",\"type\":\"address\"}],\"name\":\"addCollection\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"collectionIdByName\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"name\":\"collectionIdBySymbol\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"collections\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"addressType\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"findAgentByAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"collection\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"next\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"isKnownCollection\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"listCollections\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"moderatorCount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nextTokenId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"collectionAddress\",\"type\":\"address\"}],\"internalType\":\"structAgentCollectionInfo[]\",\"name\":\"infos\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"next\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"}],\"name\":\"moderatorOf\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"moderatorCollections\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"next\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nextCollectionId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"callerConfirmation\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// AgentCollectionsManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use AgentCollectionsManagerMetaData.ABI instead.
var AgentCollectionsManagerABI = AgentCollectionsManagerMetaData.ABI

// AgentCollectionsManager is an auto generated Go binding around an Ethereum contract.
type AgentCollectionsManager struct {
	AgentCollectionsManagerCaller     // Read-only binding to the contract
	AgentCollectionsManagerTransactor // Write-only binding to the contract
	AgentCollectionsManagerFilterer   // Log filterer for contract events
}

// AgentCollectionsManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type AgentCollectionsManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentCollectionsManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AgentCollectionsManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentCollectionsManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AgentCollectionsManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentCollectionsManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AgentCollectionsManagerSession struct {
	Contract     *AgentCollectionsManager // Generic contract binding to set the session for
	CallOpts     bind.CallOpts            // Call options to use throughout this session
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// AgentCollectionsManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AgentCollectionsManagerCallerSession struct {
	Contract *AgentCollectionsManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                  // Call options to use throughout this session
}

// AgentCollectionsManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AgentCollectionsManagerTransactorSession struct {
	Contract     *AgentCollectionsManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// AgentCollectionsManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type AgentCollectionsManagerRaw struct {
	Contract *AgentCollectionsManager // Generic contract binding to access the raw methods on
}

// AgentCollectionsManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AgentCollectionsManagerCallerRaw struct {
	Contract *AgentCollectionsManagerCaller // Generic read-only contract binding to access the raw methods on
}

// AgentCollectionsManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AgentCollectionsManagerTransactorRaw struct {
	Contract *AgentCollectionsManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAgentCollectionsManager creates a new instance of AgentCollectionsManager, bound to a specific deployed contract.
func NewAgentCollectionsManager(address common.Address, backend bind.ContractBackend) (*AgentCollectionsManager, error) {
	contract, err := bindAgentCollectionsManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AgentCollectionsManager{AgentCollectionsManagerCaller: AgentCollectionsManagerCaller{contract: contract}, AgentCollectionsManagerTransactor: AgentCollectionsManagerTransactor{contract: contract}, AgentCollectionsManagerFilterer: AgentCollectionsManagerFilterer{contract: contract}}, nil
}

// NewAgentCollectionsManagerCaller creates a new read-only instance of AgentCollectionsManager, bound to a specific deployed contract.
func NewAgentCollectionsManagerCaller(address common.Address, caller bind.ContractCaller) (*AgentCollectionsManagerCaller, error) {
	contract, err := bindAgentCollectionsManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AgentCollectionsManagerCaller{contract: contract}, nil
}

// NewAgentCollectionsManagerTransactor creates a new write-only instance of AgentCollectionsManager, bound to a specific deployed contract.
func NewAgentCollectionsManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*AgentCollectionsManagerTransactor, error) {
	contract, err := bindAgentCollectionsManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AgentCollectionsManagerTransactor{contract: contract}, nil
}

// NewAgentCollectionsManagerFilterer creates a new log filterer instance of AgentCollectionsManager, bound to a specific deployed contract.
func NewAgentCollectionsManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*AgentCollectionsManagerFilterer, error) {
	contract, err := bindAgentCollectionsManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AgentCollectionsManagerFilterer{contract: contract}, nil
}

// bindAgentCollectionsManager binds a generic wrapper to an already deployed contract.
func bindAgentCollectionsManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AgentCollectionsManagerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AgentCollectionsManager *AgentCollectionsManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AgentCollectionsManager.Contract.AgentCollectionsManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AgentCollectionsManager *AgentCollectionsManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.AgentCollectionsManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AgentCollectionsManager *AgentCollectionsManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.AgentCollectionsManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AgentCollectionsManager *AgentCollectionsManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AgentCollectionsManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AgentCollectionsManager *AgentCollectionsManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AgentCollectionsManager *AgentCollectionsManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.contract.Transact(opts, method, params...)
}

// CREATORROLE is a free data retrieval call binding the contract method 0x8aeda25a.
//
// Solidity: function CREATOR_ROLE() view returns(bytes32)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) CREATORROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "CREATOR_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CREATORROLE is a free data retrieval call binding the contract method 0x8aeda25a.
//
// Solidity: function CREATOR_ROLE() view returns(bytes32)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) CREATORROLE() ([32]byte, error) {
	return _AgentCollectionsManager.Contract.CREATORROLE(&_AgentCollectionsManager.CallOpts)
}

// CREATORROLE is a free data retrieval call binding the contract method 0x8aeda25a.
//
// Solidity: function CREATOR_ROLE() view returns(bytes32)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) CREATORROLE() ([32]byte, error) {
	return _AgentCollectionsManager.Contract.CREATORROLE(&_AgentCollectionsManager.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _AgentCollectionsManager.Contract.DEFAULTADMINROLE(&_AgentCollectionsManager.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _AgentCollectionsManager.Contract.DEFAULTADMINROLE(&_AgentCollectionsManager.CallOpts)
}

// CollectionIdByName is a free data retrieval call binding the contract method 0x08fa3c65.
//
// Solidity: function collectionIdByName(string ) view returns(uint256)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) CollectionIdByName(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "collectionIdByName", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CollectionIdByName is a free data retrieval call binding the contract method 0x08fa3c65.
//
// Solidity: function collectionIdByName(string ) view returns(uint256)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) CollectionIdByName(arg0 string) (*big.Int, error) {
	return _AgentCollectionsManager.Contract.CollectionIdByName(&_AgentCollectionsManager.CallOpts, arg0)
}

// CollectionIdByName is a free data retrieval call binding the contract method 0x08fa3c65.
//
// Solidity: function collectionIdByName(string ) view returns(uint256)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) CollectionIdByName(arg0 string) (*big.Int, error) {
	return _AgentCollectionsManager.Contract.CollectionIdByName(&_AgentCollectionsManager.CallOpts, arg0)
}

// CollectionIdBySymbol is a free data retrieval call binding the contract method 0x103b970e.
//
// Solidity: function collectionIdBySymbol(string ) view returns(uint256)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) CollectionIdBySymbol(opts *bind.CallOpts, arg0 string) (*big.Int, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "collectionIdBySymbol", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CollectionIdBySymbol is a free data retrieval call binding the contract method 0x103b970e.
//
// Solidity: function collectionIdBySymbol(string ) view returns(uint256)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) CollectionIdBySymbol(arg0 string) (*big.Int, error) {
	return _AgentCollectionsManager.Contract.CollectionIdBySymbol(&_AgentCollectionsManager.CallOpts, arg0)
}

// CollectionIdBySymbol is a free data retrieval call binding the contract method 0x103b970e.
//
// Solidity: function collectionIdBySymbol(string ) view returns(uint256)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) CollectionIdBySymbol(arg0 string) (*big.Int, error) {
	return _AgentCollectionsManager.Contract.CollectionIdBySymbol(&_AgentCollectionsManager.CallOpts, arg0)
}

// Collections is a free data retrieval call binding the contract method 0xfdbda0ec.
//
// Solidity: function collections(uint256 ) view returns(address)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) Collections(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "collections", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Collections is a free data retrieval call binding the contract method 0xfdbda0ec.
//
// Solidity: function collections(uint256 ) view returns(address)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) Collections(arg0 *big.Int) (common.Address, error) {
	return _AgentCollectionsManager.Contract.Collections(&_AgentCollectionsManager.CallOpts, arg0)
}

// Collections is a free data retrieval call binding the contract method 0xfdbda0ec.
//
// Solidity: function collections(uint256 ) view returns(address)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) Collections(arg0 *big.Int) (common.Address, error) {
	return _AgentCollectionsManager.Contract.Collections(&_AgentCollectionsManager.CallOpts, arg0)
}

// FindAgentByAddress is a free data retrieval call binding the contract method 0xe32f72a1.
//
// Solidity: function findAgentByAddress(string addressType, string value, uint256 offset, uint256 limit) view returns(address collection, uint256 tokenId, string name, uint256 next)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) FindAgentByAddress(opts *bind.CallOpts, addressType string, value string, offset *big.Int, limit *big.Int) (struct {
	Collection common.Address
	TokenId    *big.Int
	Name       string
	Next       *big.Int
}, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "findAgentByAddress", addressType, value, offset, limit)

	outstruct := new(struct {
		Collection common.Address
		TokenId    *big.Int
		Name       string
		Next       *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Collection = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.TokenId = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Name = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.Next = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// FindAgentByAddress is a free data retrieval call binding the contract method 0xe32f72a1.
//
// Solidity: function findAgentByAddress(string addressType, string value, uint256 offset, uint256 limit) view returns(address collection, uint256 tokenId, string name, uint256 next)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) FindAgentByAddress(addressType string, value string, offset *big.Int, limit *big.Int) (struct {
	Collection common.Address
	TokenId    *big.Int
	Name       string
	Next       *big.Int
}, error) {
	return _AgentCollectionsManager.Contract.FindAgentByAddress(&_AgentCollectionsManager.CallOpts, addressType, value, offset, limit)
}

// FindAgentByAddress is a free data retrieval call binding the contract method 0xe32f72a1.
//
// Solidity: function findAgentByAddress(string addressType, string value, uint256 offset, uint256 limit) view returns(address collection, uint256 tokenId, string name, uint256 next)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) FindAgentByAddress(addressType string, value string, offset *big.Int, limit *big.Int) (struct {
	Collection common.Address
	TokenId    *big.Int
	Name       string
	Next       *big.Int
}, error) {
	return _AgentCollectionsManager.Contract.FindAgentByAddress(&_AgentCollectionsManager.CallOpts, addressType, value, offset, limit)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _AgentCollectionsManager.Contract.GetRoleAdmin(&_AgentCollectionsManager.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _AgentCollectionsManager.Contract.GetRoleAdmin(&_AgentCollectionsManager.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _AgentCollectionsManager.Contract.HasRole(&_AgentCollectionsManager.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _AgentCollectionsManager.Contract.HasRole(&_AgentCollectionsManager.CallOpts, role, account)
}

// IsKnownCollection is a free data retrieval call binding the contract method 0x3637b244.
//
// Solidity: function isKnownCollection(address ) view returns(bool)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) IsKnownCollection(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "isKnownCollection", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsKnownCollection is a free data retrieval call binding the contract method 0x3637b244.
//
// Solidity: function isKnownCollection(address ) view returns(bool)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) IsKnownCollection(arg0 common.Address) (bool, error) {
	return _AgentCollectionsManager.Contract.IsKnownCollection(&_AgentCollectionsManager.CallOpts, arg0)
}

// IsKnownCollection is a free data retrieval call binding the contract method 0x3637b244.
//
// Solidity: function isKnownCollection(address ) view returns(bool)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) IsKnownCollection(arg0 common.Address) (bool, error) {
	return _AgentCollectionsManager.Contract.IsKnownCollection(&_AgentCollectionsManager.CallOpts, arg0)
}

// ListCollections is a free data retrieval call binding the contract method 0xc4c36e65.
//
// Solidity: function listCollections(uint256 offset, uint256 limit) view returns((string,string,uint256,uint256,address)[] infos, uint256 next)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) ListCollections(opts *bind.CallOpts, offset *big.Int, limit *big.Int) (struct {
	Infos []AgentCollectionInfo
	Next  *big.Int
}, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "listCollections", offset, limit)

	outstruct := new(struct {
		Infos []AgentCollectionInfo
		Next  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Infos = *abi.ConvertType(out[0], new([]AgentCollectionInfo)).(*[]AgentCollectionInfo)
	outstruct.Next = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// ListCollections is a free data retrieval call binding the contract method 0xc4c36e65.
//
// Solidity: function listCollections(uint256 offset, uint256 limit) view returns((string,string,uint256,uint256,address)[] infos, uint256 next)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) ListCollections(offset *big.Int, limit *big.Int) (struct {
	Infos []AgentCollectionInfo
	Next  *big.Int
}, error) {
	return _AgentCollectionsManager.Contract.ListCollections(&_AgentCollectionsManager.CallOpts, offset, limit)
}

// ListCollections is a free data retrieval call binding the contract method 0xc4c36e65.
//
// Solidity: function listCollections(uint256 offset, uint256 limit) view returns((string,string,uint256,uint256,address)[] infos, uint256 next)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) ListCollections(offset *big.Int, limit *big.Int) (struct {
	Infos []AgentCollectionInfo
	Next  *big.Int
}, error) {
	return _AgentCollectionsManager.Contract.ListCollections(&_AgentCollectionsManager.CallOpts, offset, limit)
}

// ModeratorOf is a free data retrieval call binding the contract method 0xc080ccab.
//
// Solidity: function moderatorOf(address account, uint256 offset, uint256 limit) view returns(address[] moderatorCollections, uint256 next)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) ModeratorOf(opts *bind.CallOpts, account common.Address, offset *big.Int, limit *big.Int) (struct {
	ModeratorCollections []common.Address
	Next                 *big.Int
}, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "moderatorOf", account, offset, limit)

	outstruct := new(struct {
		ModeratorCollections []common.Address
		Next                 *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ModeratorCollections = *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	outstruct.Next = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// ModeratorOf is a free data retrieval call binding the contract method 0xc080ccab.
//
// Solidity: function moderatorOf(address account, uint256 offset, uint256 limit) view returns(address[] moderatorCollections, uint256 next)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) ModeratorOf(account common.Address, offset *big.Int, limit *big.Int) (struct {
	ModeratorCollections []common.Address
	Next                 *big.Int
}, error) {
	return _AgentCollectionsManager.Contract.ModeratorOf(&_AgentCollectionsManager.CallOpts, account, offset, limit)
}

// ModeratorOf is a free data retrieval call binding the contract method 0xc080ccab.
//
// Solidity: function moderatorOf(address account, uint256 offset, uint256 limit) view returns(address[] moderatorCollections, uint256 next)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) ModeratorOf(account common.Address, offset *big.Int, limit *big.Int) (struct {
	ModeratorCollections []common.Address
	Next                 *big.Int
}, error) {
	return _AgentCollectionsManager.Contract.ModeratorOf(&_AgentCollectionsManager.CallOpts, account, offset, limit)
}

// NextCollectionId is a free data retrieval call binding the contract method 0xe77d6f7c.
//
// Solidity: function nextCollectionId() view returns(uint256)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) NextCollectionId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "nextCollectionId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NextCollectionId is a free data retrieval call binding the contract method 0xe77d6f7c.
//
// Solidity: function nextCollectionId() view returns(uint256)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) NextCollectionId() (*big.Int, error) {
	return _AgentCollectionsManager.Contract.NextCollectionId(&_AgentCollectionsManager.CallOpts)
}

// NextCollectionId is a free data retrieval call binding the contract method 0xe77d6f7c.
//
// Solidity: function nextCollectionId() view returns(uint256)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) NextCollectionId() (*big.Int, error) {
	return _AgentCollectionsManager.Contract.NextCollectionId(&_AgentCollectionsManager.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AgentCollectionsManager *AgentCollectionsManagerCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _AgentCollectionsManager.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AgentCollectionsManager *AgentCollectionsManagerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AgentCollectionsManager.Contract.SupportsInterface(&_AgentCollectionsManager.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AgentCollectionsManager *AgentCollectionsManagerCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AgentCollectionsManager.Contract.SupportsInterface(&_AgentCollectionsManager.CallOpts, interfaceId)
}

// AddCollection is a paid mutator transaction binding the contract method 0xa174e77a.
//
// Solidity: function addCollection(address newCollection) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerTransactor) AddCollection(opts *bind.TransactOpts, newCollection common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.contract.Transact(opts, "addCollection", newCollection)
}

// AddCollection is a paid mutator transaction binding the contract method 0xa174e77a.
//
// Solidity: function addCollection(address newCollection) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerSession) AddCollection(newCollection common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.AddCollection(&_AgentCollectionsManager.TransactOpts, newCollection)
}

// AddCollection is a paid mutator transaction binding the contract method 0xa174e77a.
//
// Solidity: function addCollection(address newCollection) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerTransactorSession) AddCollection(newCollection common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.AddCollection(&_AgentCollectionsManager.TransactOpts, newCollection)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.GrantRole(&_AgentCollectionsManager.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.GrantRole(&_AgentCollectionsManager.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.RenounceRole(&_AgentCollectionsManager.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.RenounceRole(&_AgentCollectionsManager.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.RevokeRole(&_AgentCollectionsManager.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AgentCollectionsManager *AgentCollectionsManagerTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AgentCollectionsManager.Contract.RevokeRole(&_AgentCollectionsManager.TransactOpts, role, account)
}

// AgentCollectionsManagerCollectionCreatedIterator is returned from FilterCollectionCreated and is used to iterate over the raw logs and unpacked data for CollectionCreated events raised by the AgentCollectionsManager contract.
type AgentCollectionsManagerCollectionCreatedIterator struct {
	Event *AgentCollectionsManagerCollectionCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AgentCollectionsManagerCollectionCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentCollectionsManagerCollectionCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AgentCollectionsManagerCollectionCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AgentCollectionsManagerCollectionCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentCollectionsManagerCollectionCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentCollectionsManagerCollectionCreated represents a CollectionCreated event raised by the AgentCollectionsManager contract.
type AgentCollectionsManagerCollectionCreated struct {
	CollectionId      *big.Int
	CollectionAddress common.Address
	Name              string
	Symbol            string
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterCollectionCreated is a free log retrieval operation binding the contract event 0xe1344d16944f31a1bd59b121889e600e71fcc0716ae3fe38175b7ef2db1678b3.
//
// Solidity: event CollectionCreated(uint256 indexed collectionId, address collectionAddress, string name, string symbol)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) FilterCollectionCreated(opts *bind.FilterOpts, collectionId []*big.Int) (*AgentCollectionsManagerCollectionCreatedIterator, error) {

	var collectionIdRule []interface{}
	for _, collectionIdItem := range collectionId {
		collectionIdRule = append(collectionIdRule, collectionIdItem)
	}

	logs, sub, err := _AgentCollectionsManager.contract.FilterLogs(opts, "CollectionCreated", collectionIdRule)
	if err != nil {
		return nil, err
	}
	return &AgentCollectionsManagerCollectionCreatedIterator{contract: _AgentCollectionsManager.contract, event: "CollectionCreated", logs: logs, sub: sub}, nil
}

// WatchCollectionCreated is a free log subscription operation binding the contract event 0xe1344d16944f31a1bd59b121889e600e71fcc0716ae3fe38175b7ef2db1678b3.
//
// Solidity: event CollectionCreated(uint256 indexed collectionId, address collectionAddress, string name, string symbol)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) WatchCollectionCreated(opts *bind.WatchOpts, sink chan<- *AgentCollectionsManagerCollectionCreated, collectionId []*big.Int) (event.Subscription, error) {

	var collectionIdRule []interface{}
	for _, collectionIdItem := range collectionId {
		collectionIdRule = append(collectionIdRule, collectionIdItem)
	}

	logs, sub, err := _AgentCollectionsManager.contract.WatchLogs(opts, "CollectionCreated", collectionIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentCollectionsManagerCollectionCreated)
				if err := _AgentCollectionsManager.contract.UnpackLog(event, "CollectionCreated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCollectionCreated is a log parse operation binding the contract event 0xe1344d16944f31a1bd59b121889e600e71fcc0716ae3fe38175b7ef2db1678b3.
//
// Solidity: event CollectionCreated(uint256 indexed collectionId, address collectionAddress, string name, string symbol)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) ParseCollectionCreated(log types.Log) (*AgentCollectionsManagerCollectionCreated, error) {
	event := new(AgentCollectionsManagerCollectionCreated)
	if err := _AgentCollectionsManager.contract.UnpackLog(event, "CollectionCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentCollectionsManagerRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the AgentCollectionsManager contract.
type AgentCollectionsManagerRoleAdminChangedIterator struct {
	Event *AgentCollectionsManagerRoleAdminChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AgentCollectionsManagerRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentCollectionsManagerRoleAdminChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AgentCollectionsManagerRoleAdminChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AgentCollectionsManagerRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentCollectionsManagerRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentCollectionsManagerRoleAdminChanged represents a RoleAdminChanged event raised by the AgentCollectionsManager contract.
type AgentCollectionsManagerRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*AgentCollectionsManagerRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _AgentCollectionsManager.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &AgentCollectionsManagerRoleAdminChangedIterator{contract: _AgentCollectionsManager.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *AgentCollectionsManagerRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _AgentCollectionsManager.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentCollectionsManagerRoleAdminChanged)
				if err := _AgentCollectionsManager.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) ParseRoleAdminChanged(log types.Log) (*AgentCollectionsManagerRoleAdminChanged, error) {
	event := new(AgentCollectionsManagerRoleAdminChanged)
	if err := _AgentCollectionsManager.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentCollectionsManagerRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the AgentCollectionsManager contract.
type AgentCollectionsManagerRoleGrantedIterator struct {
	Event *AgentCollectionsManagerRoleGranted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AgentCollectionsManagerRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentCollectionsManagerRoleGranted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AgentCollectionsManagerRoleGranted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AgentCollectionsManagerRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentCollectionsManagerRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentCollectionsManagerRoleGranted represents a RoleGranted event raised by the AgentCollectionsManager contract.
type AgentCollectionsManagerRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AgentCollectionsManagerRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AgentCollectionsManager.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AgentCollectionsManagerRoleGrantedIterator{contract: _AgentCollectionsManager.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *AgentCollectionsManagerRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AgentCollectionsManager.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentCollectionsManagerRoleGranted)
				if err := _AgentCollectionsManager.contract.UnpackLog(event, "RoleGranted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) ParseRoleGranted(log types.Log) (*AgentCollectionsManagerRoleGranted, error) {
	event := new(AgentCollectionsManagerRoleGranted)
	if err := _AgentCollectionsManager.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentCollectionsManagerRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the AgentCollectionsManager contract.
type AgentCollectionsManagerRoleRevokedIterator struct {
	Event *AgentCollectionsManagerRoleRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AgentCollectionsManagerRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentCollectionsManagerRoleRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AgentCollectionsManagerRoleRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AgentCollectionsManagerRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentCollectionsManagerRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentCollectionsManagerRoleRevoked represents a RoleRevoked event raised by the AgentCollectionsManager contract.
type AgentCollectionsManagerRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AgentCollectionsManagerRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AgentCollectionsManager.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AgentCollectionsManagerRoleRevokedIterator{contract: _AgentCollectionsManager.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *AgentCollectionsManagerRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AgentCollectionsManager.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentCollectionsManagerRoleRevoked)
				if err := _AgentCollectionsManager.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AgentCollectionsManager *AgentCollectionsManagerFilterer) ParseRoleRevoked(log types.Log) (*AgentCollectionsManagerRoleRevoked, error) {
	event := new(AgentCollectionsManagerRoleRevoked)
	if err := _AgentCollectionsManager.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// --- RawLog() methods added by go generate ---
func (e AgentCollectionsManagerCollectionCreated) RawLog() types.Log { return e.Raw }
func (e AgentCollectionsManagerRoleAdminChanged) RawLog() types.Log { return e.Raw }
func (e AgentCollectionsManagerRoleGranted) RawLog() types.Log { return e.Raw }
func (e AgentCollectionsManagerRoleRevoked) RawLog() types.Log { return e.Raw }
