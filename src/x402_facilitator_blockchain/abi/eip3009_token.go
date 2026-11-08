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

// Eip3009TokenMetaData contains all meta data concerning the Eip3009Token contract.
var Eip3009TokenMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"balanceOf\",\"stateMutability\":\"view\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}]},{\"type\":\"function\",\"name\":\"authorizationState\",\"stateMutability\":\"view\",\"inputs\":[{\"name\":\"authorizer\",\"type\":\"address\"},{\"name\":\"nonce\",\"type\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}]},{\"type\":\"function\",\"name\":\"transferWithAuthorization\",\"stateMutability\":\"nonpayable\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"},{\"name\":\"validAfter\",\"type\":\"uint256\"},{\"name\":\"validBefore\",\"type\":\"uint256\"},{\"name\":\"nonce\",\"type\":\"bytes32\"},{\"name\":\"v\",\"type\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\"}],\"outputs\":[]}]",
}

// Eip3009TokenABI is the input ABI used to generate the binding from.
// Deprecated: Use Eip3009TokenMetaData.ABI instead.
var Eip3009TokenABI = Eip3009TokenMetaData.ABI

// Eip3009Token is an auto generated Go binding around an Ethereum contract.
type Eip3009Token struct {
	Eip3009TokenCaller     // Read-only binding to the contract
	Eip3009TokenTransactor // Write-only binding to the contract
	Eip3009TokenFilterer   // Log filterer for contract events
}

// Eip3009TokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type Eip3009TokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Eip3009TokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type Eip3009TokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Eip3009TokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type Eip3009TokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Eip3009TokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type Eip3009TokenSession struct {
	Contract     *Eip3009Token     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Eip3009TokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type Eip3009TokenCallerSession struct {
	Contract *Eip3009TokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// Eip3009TokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type Eip3009TokenTransactorSession struct {
	Contract     *Eip3009TokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// Eip3009TokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type Eip3009TokenRaw struct {
	Contract *Eip3009Token // Generic contract binding to access the raw methods on
}

// Eip3009TokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type Eip3009TokenCallerRaw struct {
	Contract *Eip3009TokenCaller // Generic read-only contract binding to access the raw methods on
}

// Eip3009TokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type Eip3009TokenTransactorRaw struct {
	Contract *Eip3009TokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEip3009Token creates a new instance of Eip3009Token, bound to a specific deployed contract.
func NewEip3009Token(address common.Address, backend bind.ContractBackend) (*Eip3009Token, error) {
	contract, err := bindEip3009Token(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Eip3009Token{Eip3009TokenCaller: Eip3009TokenCaller{contract: contract}, Eip3009TokenTransactor: Eip3009TokenTransactor{contract: contract}, Eip3009TokenFilterer: Eip3009TokenFilterer{contract: contract}}, nil
}

// NewEip3009TokenCaller creates a new read-only instance of Eip3009Token, bound to a specific deployed contract.
func NewEip3009TokenCaller(address common.Address, caller bind.ContractCaller) (*Eip3009TokenCaller, error) {
	contract, err := bindEip3009Token(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Eip3009TokenCaller{contract: contract}, nil
}

// NewEip3009TokenTransactor creates a new write-only instance of Eip3009Token, bound to a specific deployed contract.
func NewEip3009TokenTransactor(address common.Address, transactor bind.ContractTransactor) (*Eip3009TokenTransactor, error) {
	contract, err := bindEip3009Token(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &Eip3009TokenTransactor{contract: contract}, nil
}

// NewEip3009TokenFilterer creates a new log filterer instance of Eip3009Token, bound to a specific deployed contract.
func NewEip3009TokenFilterer(address common.Address, filterer bind.ContractFilterer) (*Eip3009TokenFilterer, error) {
	contract, err := bindEip3009Token(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &Eip3009TokenFilterer{contract: contract}, nil
}

// bindEip3009Token binds a generic wrapper to an already deployed contract.
func bindEip3009Token(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := Eip3009TokenMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Eip3009Token *Eip3009TokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Eip3009Token.Contract.Eip3009TokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Eip3009Token *Eip3009TokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Eip3009Token.Contract.Eip3009TokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Eip3009Token *Eip3009TokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Eip3009Token.Contract.Eip3009TokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Eip3009Token *Eip3009TokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Eip3009Token.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Eip3009Token *Eip3009TokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Eip3009Token.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Eip3009Token *Eip3009TokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Eip3009Token.Contract.contract.Transact(opts, method, params...)
}

// AuthorizationState is a free data retrieval call binding the contract method 0xe94a0102.
//
// Solidity: function authorizationState(address authorizer, bytes32 nonce) view returns(bool)
func (_Eip3009Token *Eip3009TokenCaller) AuthorizationState(opts *bind.CallOpts, authorizer common.Address, nonce [32]byte) (bool, error) {
	var out []interface{}
	err := _Eip3009Token.contract.Call(opts, &out, "authorizationState", authorizer, nonce)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AuthorizationState is a free data retrieval call binding the contract method 0xe94a0102.
//
// Solidity: function authorizationState(address authorizer, bytes32 nonce) view returns(bool)
func (_Eip3009Token *Eip3009TokenSession) AuthorizationState(authorizer common.Address, nonce [32]byte) (bool, error) {
	return _Eip3009Token.Contract.AuthorizationState(&_Eip3009Token.CallOpts, authorizer, nonce)
}

// AuthorizationState is a free data retrieval call binding the contract method 0xe94a0102.
//
// Solidity: function authorizationState(address authorizer, bytes32 nonce) view returns(bool)
func (_Eip3009Token *Eip3009TokenCallerSession) AuthorizationState(authorizer common.Address, nonce [32]byte) (bool, error) {
	return _Eip3009Token.Contract.AuthorizationState(&_Eip3009Token.CallOpts, authorizer, nonce)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Eip3009Token *Eip3009TokenCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Eip3009Token.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Eip3009Token *Eip3009TokenSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _Eip3009Token.Contract.BalanceOf(&_Eip3009Token.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_Eip3009Token *Eip3009TokenCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _Eip3009Token.Contract.BalanceOf(&_Eip3009Token.CallOpts, account)
}

// TransferWithAuthorization is a paid mutator transaction binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address from, address to, uint256 value, uint256 validAfter, uint256 validBefore, bytes32 nonce, uint8 v, bytes32 r, bytes32 s) returns()
func (_Eip3009Token *Eip3009TokenTransactor) TransferWithAuthorization(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int, validAfter *big.Int, validBefore *big.Int, nonce [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Eip3009Token.contract.Transact(opts, "transferWithAuthorization", from, to, value, validAfter, validBefore, nonce, v, r, s)
}

// TransferWithAuthorization is a paid mutator transaction binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address from, address to, uint256 value, uint256 validAfter, uint256 validBefore, bytes32 nonce, uint8 v, bytes32 r, bytes32 s) returns()
func (_Eip3009Token *Eip3009TokenSession) TransferWithAuthorization(from common.Address, to common.Address, value *big.Int, validAfter *big.Int, validBefore *big.Int, nonce [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Eip3009Token.Contract.TransferWithAuthorization(&_Eip3009Token.TransactOpts, from, to, value, validAfter, validBefore, nonce, v, r, s)
}

// TransferWithAuthorization is a paid mutator transaction binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address from, address to, uint256 value, uint256 validAfter, uint256 validBefore, bytes32 nonce, uint8 v, bytes32 r, bytes32 s) returns()
func (_Eip3009Token *Eip3009TokenTransactorSession) TransferWithAuthorization(from common.Address, to common.Address, value *big.Int, validAfter *big.Int, validBefore *big.Int, nonce [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Eip3009Token.Contract.TransferWithAuthorization(&_Eip3009Token.TransactOpts, from, to, value, validAfter, validBefore, nonce, v, r, s)
}
