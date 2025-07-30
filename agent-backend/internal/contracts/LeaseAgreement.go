// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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

// LeaseAgreementLease is an auto generated low-level Go binding around an user-defined struct.
type LeaseAgreementLease struct {
	Spender       common.Address
	Earner        common.Address
	DataProductId [32]byte
	Price         *big.Int
	MaxPrice      *big.Int
	IsApproved    bool
	IsExecuted    bool
	IsDisputed    bool
	CreatedAt     *big.Int
}

// LeaseAgreementMetaData contains all meta data concerning the LeaseAgreement contract.
var LeaseAgreementMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"leaseId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"earner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"}],\"name\":\"LeaseCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"leaseId\",\"type\":\"bytes32\"}],\"name\":\"LeaseApproved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"leaseId\",\"type\":\"bytes32\"}],\"name\":\"LeaseExecuted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"earner\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"dataProductId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"maxPrice\",\"type\":\"uint256\"}],\"name\":\"createLease\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"leaseId\",\"type\":\"bytes32\"}],\"name\":\"approveLease\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"leaseId\",\"type\":\"bytes32\"}],\"name\":\"executeLease\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"leaseId\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"name\":\"raiseDispute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"leaseId\",\"type\":\"bytes32\"}],\"name\":\"getLease\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"earner\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"dataProductId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPrice\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isApproved\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"isExecuted\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"isDisputed\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"}],\"internalType\":\"structLeaseAgreement.Lease\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newMinPrice\",\"type\":\"uint256\"}],\"name\":\"updateMinPrice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"emergencyPause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_PRICE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"leases\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"earner\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"dataProductId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxPrice\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isApproved\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"isExecuted\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"isDisputed\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"leaseExists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// LeaseAgreementABI is the input ABI used to generate the binding from.
// Deprecated: Use LeaseAgreementMetaData.ABI instead.
var LeaseAgreementABI = LeaseAgreementMetaData.ABI

// LeaseAgreement is an auto generated Go binding around an Ethereum contract.
type LeaseAgreement struct {
	LeaseAgreementCaller     // Read-only binding to the contract
	LeaseAgreementTransactor // Write-only binding to the contract
	LeaseAgreementFilterer   // Log filterer for contract events
}

// LeaseAgreementCaller is an auto generated read-only Go binding around an Ethereum contract.
type LeaseAgreementCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LeaseAgreementTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LeaseAgreementTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LeaseAgreementFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LeaseAgreementFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LeaseAgreementSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LeaseAgreementSession struct {
	Contract     *LeaseAgreement   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// LeaseAgreementCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LeaseAgreementCallerSession struct {
	Contract *LeaseAgreementCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// LeaseAgreementTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LeaseAgreementTransactorSession struct {
	Contract     *LeaseAgreementTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// LeaseAgreementRaw is an auto generated low-level Go binding around an Ethereum contract.
type LeaseAgreementRaw struct {
	Contract *LeaseAgreement // Generic contract binding to access the raw methods on
}

// LeaseAgreementCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LeaseAgreementCallerRaw struct {
	Contract *LeaseAgreementCaller // Generic read-only contract binding to access the raw methods on
}

// LeaseAgreementTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LeaseAgreementTransactorRaw struct {
	Contract *LeaseAgreementTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLeaseAgreement creates a new instance of LeaseAgreement, bound to a specific deployed contract.
func NewLeaseAgreement(address common.Address, backend bind.ContractBackend) (*LeaseAgreement, error) {
	contract, err := bindLeaseAgreement(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &LeaseAgreement{LeaseAgreementCaller: LeaseAgreementCaller{contract: contract}, LeaseAgreementTransactor: LeaseAgreementTransactor{contract: contract}, LeaseAgreementFilterer: LeaseAgreementFilterer{contract: contract}}, nil
}

// NewLeaseAgreementCaller creates a new read-only instance of LeaseAgreement, bound to a specific deployed contract.
func NewLeaseAgreementCaller(address common.Address, caller bind.ContractCaller) (*LeaseAgreementCaller, error) {
	contract, err := bindLeaseAgreement(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LeaseAgreementCaller{contract: contract}, nil
}

// NewLeaseAgreementTransactor creates a new write-only instance of LeaseAgreement, bound to a specific deployed contract.
func NewLeaseAgreementTransactor(address common.Address, transactor bind.ContractTransactor) (*LeaseAgreementTransactor, error) {
	contract, err := bindLeaseAgreement(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LeaseAgreementTransactor{contract: contract}, nil
}

// NewLeaseAgreementFilterer creates a new log filterer instance of LeaseAgreement, bound to a specific deployed contract.
func NewLeaseAgreementFilterer(address common.Address, filterer bind.ContractFilterer) (*LeaseAgreementFilterer, error) {
	contract, err := bindLeaseAgreement(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LeaseAgreementFilterer{contract: contract}, nil
}

// bindLeaseAgreement binds a generic wrapper to an already deployed contract.
func bindLeaseAgreement(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := LeaseAgreementMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LeaseAgreement *LeaseAgreementRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _LeaseAgreement.Contract.LeaseAgreementCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LeaseAgreement *LeaseAgreementRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.LeaseAgreementTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LeaseAgreement *LeaseAgreementRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.LeaseAgreementTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LeaseAgreement *LeaseAgreementCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _LeaseAgreement.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LeaseAgreement *LeaseAgreementTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LeaseAgreement *LeaseAgreementTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.contract.Transact(opts, method, params...)
}

// MINPRICE is a free data retrieval call binding the contract method 0xad9f20a6.
//
// Solidity: function MIN_PRICE() view returns(uint256)
func (_LeaseAgreement *LeaseAgreementCaller) MINPRICE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LeaseAgreement.contract.Call(opts, &out, "MIN_PRICE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINPRICE is a free data retrieval call binding the contract method 0xad9f20a6.
//
// Solidity: function MIN_PRICE() view returns(uint256)
func (_LeaseAgreement *LeaseAgreementSession) MINPRICE() (*big.Int, error) {
	return _LeaseAgreement.Contract.MINPRICE(&_LeaseAgreement.CallOpts)
}

// MINPRICE is a free data retrieval call binding the contract method 0xad9f20a6.
//
// Solidity: function MIN_PRICE() view returns(uint256)
func (_LeaseAgreement *LeaseAgreementCallerSession) MINPRICE() (*big.Int, error) {
	return _LeaseAgreement.Contract.MINPRICE(&_LeaseAgreement.CallOpts)
}

// GetLease is a free data retrieval call binding the contract method 0x2d54ad30.
//
// Solidity: function getLease(bytes32 leaseId) view returns((address,address,bytes32,uint256,uint256,bool,bool,bool,uint256))
func (_LeaseAgreement *LeaseAgreementCaller) GetLease(opts *bind.CallOpts, leaseId [32]byte) (LeaseAgreementLease, error) {
	var out []interface{}
	err := _LeaseAgreement.contract.Call(opts, &out, "getLease", leaseId)

	if err != nil {
		return *new(LeaseAgreementLease), err
	}

	out0 := *abi.ConvertType(out[0], new(LeaseAgreementLease)).(*LeaseAgreementLease)

	return out0, err

}

// GetLease is a free data retrieval call binding the contract method 0x2d54ad30.
//
// Solidity: function getLease(bytes32 leaseId) view returns((address,address,bytes32,uint256,uint256,bool,bool,bool,uint256))
func (_LeaseAgreement *LeaseAgreementSession) GetLease(leaseId [32]byte) (LeaseAgreementLease, error) {
	return _LeaseAgreement.Contract.GetLease(&_LeaseAgreement.CallOpts, leaseId)
}

// GetLease is a free data retrieval call binding the contract method 0x2d54ad30.
//
// Solidity: function getLease(bytes32 leaseId) view returns((address,address,bytes32,uint256,uint256,bool,bool,bool,uint256))
func (_LeaseAgreement *LeaseAgreementCallerSession) GetLease(leaseId [32]byte) (LeaseAgreementLease, error) {
	return _LeaseAgreement.Contract.GetLease(&_LeaseAgreement.CallOpts, leaseId)
}

// LeaseExists is a free data retrieval call binding the contract method 0x39056e4e.
//
// Solidity: function leaseExists(bytes32 ) view returns(bool)
func (_LeaseAgreement *LeaseAgreementCaller) LeaseExists(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _LeaseAgreement.contract.Call(opts, &out, "leaseExists", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// LeaseExists is a free data retrieval call binding the contract method 0x39056e4e.
//
// Solidity: function leaseExists(bytes32 ) view returns(bool)
func (_LeaseAgreement *LeaseAgreementSession) LeaseExists(arg0 [32]byte) (bool, error) {
	return _LeaseAgreement.Contract.LeaseExists(&_LeaseAgreement.CallOpts, arg0)
}

// LeaseExists is a free data retrieval call binding the contract method 0x39056e4e.
//
// Solidity: function leaseExists(bytes32 ) view returns(bool)
func (_LeaseAgreement *LeaseAgreementCallerSession) LeaseExists(arg0 [32]byte) (bool, error) {
	return _LeaseAgreement.Contract.LeaseExists(&_LeaseAgreement.CallOpts, arg0)
}

// Leases is a free data retrieval call binding the contract method 0x1839a5a3.
//
// Solidity: function leases(bytes32 ) view returns(address spender, address earner, bytes32 dataProductId, uint256 price, uint256 maxPrice, bool isApproved, bool isExecuted, bool isDisputed, uint256 createdAt)
func (_LeaseAgreement *LeaseAgreementCaller) Leases(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Spender       common.Address
	Earner        common.Address
	DataProductId [32]byte
	Price         *big.Int
	MaxPrice      *big.Int
	IsApproved    bool
	IsExecuted    bool
	IsDisputed    bool
	CreatedAt     *big.Int
}, error) {
	var out []interface{}
	err := _LeaseAgreement.contract.Call(opts, &out, "leases", arg0)

	outstruct := new(struct {
		Spender       common.Address
		Earner        common.Address
		DataProductId [32]byte
		Price         *big.Int
		MaxPrice      *big.Int
		IsApproved    bool
		IsExecuted    bool
		IsDisputed    bool
		CreatedAt     *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Spender = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Earner = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.DataProductId = *abi.ConvertType(out[2], new([32]byte)).(*[32]byte)
	outstruct.Price = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.MaxPrice = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.IsApproved = *abi.ConvertType(out[5], new(bool)).(*bool)
	outstruct.IsExecuted = *abi.ConvertType(out[6], new(bool)).(*bool)
	outstruct.IsDisputed = *abi.ConvertType(out[7], new(bool)).(*bool)
	outstruct.CreatedAt = *abi.ConvertType(out[8], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Leases is a free data retrieval call binding the contract method 0x1839a5a3.
//
// Solidity: function leases(bytes32 ) view returns(address spender, address earner, bytes32 dataProductId, uint256 price, uint256 maxPrice, bool isApproved, bool isExecuted, bool isDisputed, uint256 createdAt)
func (_LeaseAgreement *LeaseAgreementSession) Leases(arg0 [32]byte) (struct {
	Spender       common.Address
	Earner        common.Address
	DataProductId [32]byte
	Price         *big.Int
	MaxPrice      *big.Int
	IsApproved    bool
	IsExecuted    bool
	IsDisputed    bool
	CreatedAt     *big.Int
}, error) {
	return _LeaseAgreement.Contract.Leases(&_LeaseAgreement.CallOpts, arg0)
}

// Leases is a free data retrieval call binding the contract method 0x1839a5a3.
//
// Solidity: function leases(bytes32 ) view returns(address spender, address earner, bytes32 dataProductId, uint256 price, uint256 maxPrice, bool isApproved, bool isExecuted, bool isDisputed, uint256 createdAt)
func (_LeaseAgreement *LeaseAgreementCallerSession) Leases(arg0 [32]byte) (struct {
	Spender       common.Address
	Earner        common.Address
	DataProductId [32]byte
	Price         *big.Int
	MaxPrice      *big.Int
	IsApproved    bool
	IsExecuted    bool
	IsDisputed    bool
	CreatedAt     *big.Int
}, error) {
	return _LeaseAgreement.Contract.Leases(&_LeaseAgreement.CallOpts, arg0)
}

// ApproveLease is a paid mutator transaction binding the contract method 0x9657e610.
//
// Solidity: function approveLease(bytes32 leaseId) returns()
func (_LeaseAgreement *LeaseAgreementTransactor) ApproveLease(opts *bind.TransactOpts, leaseId [32]byte) (*types.Transaction, error) {
	return _LeaseAgreement.contract.Transact(opts, "approveLease", leaseId)
}

// ApproveLease is a paid mutator transaction binding the contract method 0x9657e610.
//
// Solidity: function approveLease(bytes32 leaseId) returns()
func (_LeaseAgreement *LeaseAgreementSession) ApproveLease(leaseId [32]byte) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.ApproveLease(&_LeaseAgreement.TransactOpts, leaseId)
}

// ApproveLease is a paid mutator transaction binding the contract method 0x9657e610.
//
// Solidity: function approveLease(bytes32 leaseId) returns()
func (_LeaseAgreement *LeaseAgreementTransactorSession) ApproveLease(leaseId [32]byte) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.ApproveLease(&_LeaseAgreement.TransactOpts, leaseId)
}

// CreateLease is a paid mutator transaction binding the contract method 0x0971b7a5.
//
// Solidity: function createLease(address earner, bytes32 dataProductId, uint256 maxPrice) payable returns()
func (_LeaseAgreement *LeaseAgreementTransactor) CreateLease(opts *bind.TransactOpts, earner common.Address, dataProductId [32]byte, maxPrice *big.Int) (*types.Transaction, error) {
	return _LeaseAgreement.contract.Transact(opts, "createLease", earner, dataProductId, maxPrice)
}

// CreateLease is a paid mutator transaction binding the contract method 0x0971b7a5.
//
// Solidity: function createLease(address earner, bytes32 dataProductId, uint256 maxPrice) payable returns()
func (_LeaseAgreement *LeaseAgreementSession) CreateLease(earner common.Address, dataProductId [32]byte, maxPrice *big.Int) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.CreateLease(&_LeaseAgreement.TransactOpts, earner, dataProductId, maxPrice)
}

// CreateLease is a paid mutator transaction binding the contract method 0x0971b7a5.
//
// Solidity: function createLease(address earner, bytes32 dataProductId, uint256 maxPrice) payable returns()
func (_LeaseAgreement *LeaseAgreementTransactorSession) CreateLease(earner common.Address, dataProductId [32]byte, maxPrice *big.Int) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.CreateLease(&_LeaseAgreement.TransactOpts, earner, dataProductId, maxPrice)
}

// EmergencyPause is a paid mutator transaction binding the contract method 0x51858e27.
//
// Solidity: function emergencyPause() returns()
func (_LeaseAgreement *LeaseAgreementTransactor) EmergencyPause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LeaseAgreement.contract.Transact(opts, "emergencyPause")
}

// EmergencyPause is a paid mutator transaction binding the contract method 0x51858e27.
//
// Solidity: function emergencyPause() returns()
func (_LeaseAgreement *LeaseAgreementSession) EmergencyPause() (*types.Transaction, error) {
	return _LeaseAgreement.Contract.EmergencyPause(&_LeaseAgreement.TransactOpts)
}

// EmergencyPause is a paid mutator transaction binding the contract method 0x51858e27.
//
// Solidity: function emergencyPause() returns()
func (_LeaseAgreement *LeaseAgreementTransactorSession) EmergencyPause() (*types.Transaction, error) {
	return _LeaseAgreement.Contract.EmergencyPause(&_LeaseAgreement.TransactOpts)
}

// ExecuteLease is a paid mutator transaction binding the contract method 0x71f05382.
//
// Solidity: function executeLease(bytes32 leaseId) returns()
func (_LeaseAgreement *LeaseAgreementTransactor) ExecuteLease(opts *bind.TransactOpts, leaseId [32]byte) (*types.Transaction, error) {
	return _LeaseAgreement.contract.Transact(opts, "executeLease", leaseId)
}

// ExecuteLease is a paid mutator transaction binding the contract method 0x71f05382.
//
// Solidity: function executeLease(bytes32 leaseId) returns()
func (_LeaseAgreement *LeaseAgreementSession) ExecuteLease(leaseId [32]byte) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.ExecuteLease(&_LeaseAgreement.TransactOpts, leaseId)
}

// ExecuteLease is a paid mutator transaction binding the contract method 0x71f05382.
//
// Solidity: function executeLease(bytes32 leaseId) returns()
func (_LeaseAgreement *LeaseAgreementTransactorSession) ExecuteLease(leaseId [32]byte) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.ExecuteLease(&_LeaseAgreement.TransactOpts, leaseId)
}

// RaiseDispute is a paid mutator transaction binding the contract method 0xbe7b8a77.
//
// Solidity: function raiseDispute(bytes32 leaseId, string reason) returns()
func (_LeaseAgreement *LeaseAgreementTransactor) RaiseDispute(opts *bind.TransactOpts, leaseId [32]byte, reason string) (*types.Transaction, error) {
	return _LeaseAgreement.contract.Transact(opts, "raiseDispute", leaseId, reason)
}

// RaiseDispute is a paid mutator transaction binding the contract method 0xbe7b8a77.
//
// Solidity: function raiseDispute(bytes32 leaseId, string reason) returns()
func (_LeaseAgreement *LeaseAgreementSession) RaiseDispute(leaseId [32]byte, reason string) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.RaiseDispute(&_LeaseAgreement.TransactOpts, leaseId, reason)
}

// RaiseDispute is a paid mutator transaction binding the contract method 0xbe7b8a77.
//
// Solidity: function raiseDispute(bytes32 leaseId, string reason) returns()
func (_LeaseAgreement *LeaseAgreementTransactorSession) RaiseDispute(leaseId [32]byte, reason string) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.RaiseDispute(&_LeaseAgreement.TransactOpts, leaseId, reason)
}

// UpdateMinPrice is a paid mutator transaction binding the contract method 0x0539fcd1.
//
// Solidity: function updateMinPrice(uint256 newMinPrice) returns()
func (_LeaseAgreement *LeaseAgreementTransactor) UpdateMinPrice(opts *bind.TransactOpts, newMinPrice *big.Int) (*types.Transaction, error) {
	return _LeaseAgreement.contract.Transact(opts, "updateMinPrice", newMinPrice)
}

// UpdateMinPrice is a paid mutator transaction binding the contract method 0x0539fcd1.
//
// Solidity: function updateMinPrice(uint256 newMinPrice) returns()
func (_LeaseAgreement *LeaseAgreementSession) UpdateMinPrice(newMinPrice *big.Int) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.UpdateMinPrice(&_LeaseAgreement.TransactOpts, newMinPrice)
}

// UpdateMinPrice is a paid mutator transaction binding the contract method 0x0539fcd1.
//
// Solidity: function updateMinPrice(uint256 newMinPrice) returns()
func (_LeaseAgreement *LeaseAgreementTransactorSession) UpdateMinPrice(newMinPrice *big.Int) (*types.Transaction, error) {
	return _LeaseAgreement.Contract.UpdateMinPrice(&_LeaseAgreement.TransactOpts, newMinPrice)
}

// LeaseAgreementLeaseApprovedIterator is returned from FilterLeaseApproved and is used to iterate over the raw logs and unpacked data for LeaseApproved events raised by the LeaseAgreement contract.
type LeaseAgreementLeaseApprovedIterator struct {
	Event *LeaseAgreementLeaseApproved // Event containing the contract specifics and raw log

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
func (it *LeaseAgreementLeaseApprovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LeaseAgreementLeaseApproved)
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
		it.Event = new(LeaseAgreementLeaseApproved)
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
func (it *LeaseAgreementLeaseApprovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LeaseAgreementLeaseApprovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LeaseAgreementLeaseApproved represents a LeaseApproved event raised by the LeaseAgreement contract.
type LeaseAgreementLeaseApproved struct {
	LeaseId [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterLeaseApproved is a free log retrieval operation binding the contract event 0xdc60d58f59485d15206729655c62ed7264bf00381216c4aac797c6e9456c5098.
//
// Solidity: event LeaseApproved(bytes32 indexed leaseId)
func (_LeaseAgreement *LeaseAgreementFilterer) FilterLeaseApproved(opts *bind.FilterOpts, leaseId [][32]byte) (*LeaseAgreementLeaseApprovedIterator, error) {

	var leaseIdRule []interface{}
	for _, leaseIdItem := range leaseId {
		leaseIdRule = append(leaseIdRule, leaseIdItem)
	}

	logs, sub, err := _LeaseAgreement.contract.FilterLogs(opts, "LeaseApproved", leaseIdRule)
	if err != nil {
		return nil, err
	}
	return &LeaseAgreementLeaseApprovedIterator{contract: _LeaseAgreement.contract, event: "LeaseApproved", logs: logs, sub: sub}, nil
}

// WatchLeaseApproved is a free log subscription operation binding the contract event 0xdc60d58f59485d15206729655c62ed7264bf00381216c4aac797c6e9456c5098.
//
// Solidity: event LeaseApproved(bytes32 indexed leaseId)
func (_LeaseAgreement *LeaseAgreementFilterer) WatchLeaseApproved(opts *bind.WatchOpts, sink chan<- *LeaseAgreementLeaseApproved, leaseId [][32]byte) (event.Subscription, error) {

	var leaseIdRule []interface{}
	for _, leaseIdItem := range leaseId {
		leaseIdRule = append(leaseIdRule, leaseIdItem)
	}

	logs, sub, err := _LeaseAgreement.contract.WatchLogs(opts, "LeaseApproved", leaseIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LeaseAgreementLeaseApproved)
				if err := _LeaseAgreement.contract.UnpackLog(event, "LeaseApproved", log); err != nil {
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

// ParseLeaseApproved is a log parse operation binding the contract event 0xdc60d58f59485d15206729655c62ed7264bf00381216c4aac797c6e9456c5098.
//
// Solidity: event LeaseApproved(bytes32 indexed leaseId)
func (_LeaseAgreement *LeaseAgreementFilterer) ParseLeaseApproved(log types.Log) (*LeaseAgreementLeaseApproved, error) {
	event := new(LeaseAgreementLeaseApproved)
	if err := _LeaseAgreement.contract.UnpackLog(event, "LeaseApproved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LeaseAgreementLeaseCreatedIterator is returned from FilterLeaseCreated and is used to iterate over the raw logs and unpacked data for LeaseCreated events raised by the LeaseAgreement contract.
type LeaseAgreementLeaseCreatedIterator struct {
	Event *LeaseAgreementLeaseCreated // Event containing the contract specifics and raw log

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
func (it *LeaseAgreementLeaseCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LeaseAgreementLeaseCreated)
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
		it.Event = new(LeaseAgreementLeaseCreated)
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
func (it *LeaseAgreementLeaseCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LeaseAgreementLeaseCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LeaseAgreementLeaseCreated represents a LeaseCreated event raised by the LeaseAgreement contract.
type LeaseAgreementLeaseCreated struct {
	LeaseId [32]byte
	Spender common.Address
	Earner  common.Address
	Price   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterLeaseCreated is a free log retrieval operation binding the contract event 0x57acec708e7a7b4437dd4bb623ed53257e444ac360b3fb2ddf9f157ad13e98a0.
//
// Solidity: event LeaseCreated(bytes32 indexed leaseId, address indexed spender, address indexed earner, uint256 price)
func (_LeaseAgreement *LeaseAgreementFilterer) FilterLeaseCreated(opts *bind.FilterOpts, leaseId [][32]byte, spender []common.Address, earner []common.Address) (*LeaseAgreementLeaseCreatedIterator, error) {

	var leaseIdRule []interface{}
	for _, leaseIdItem := range leaseId {
		leaseIdRule = append(leaseIdRule, leaseIdItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}
	var earnerRule []interface{}
	for _, earnerItem := range earner {
		earnerRule = append(earnerRule, earnerItem)
	}

	logs, sub, err := _LeaseAgreement.contract.FilterLogs(opts, "LeaseCreated", leaseIdRule, spenderRule, earnerRule)
	if err != nil {
		return nil, err
	}
	return &LeaseAgreementLeaseCreatedIterator{contract: _LeaseAgreement.contract, event: "LeaseCreated", logs: logs, sub: sub}, nil
}

// WatchLeaseCreated is a free log subscription operation binding the contract event 0x57acec708e7a7b4437dd4bb623ed53257e444ac360b3fb2ddf9f157ad13e98a0.
//
// Solidity: event LeaseCreated(bytes32 indexed leaseId, address indexed spender, address indexed earner, uint256 price)
func (_LeaseAgreement *LeaseAgreementFilterer) WatchLeaseCreated(opts *bind.WatchOpts, sink chan<- *LeaseAgreementLeaseCreated, leaseId [][32]byte, spender []common.Address, earner []common.Address) (event.Subscription, error) {

	var leaseIdRule []interface{}
	for _, leaseIdItem := range leaseId {
		leaseIdRule = append(leaseIdRule, leaseIdItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}
	var earnerRule []interface{}
	for _, earnerItem := range earner {
		earnerRule = append(earnerRule, earnerItem)
	}

	logs, sub, err := _LeaseAgreement.contract.WatchLogs(opts, "LeaseCreated", leaseIdRule, spenderRule, earnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LeaseAgreementLeaseCreated)
				if err := _LeaseAgreement.contract.UnpackLog(event, "LeaseCreated", log); err != nil {
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

// ParseLeaseCreated is a log parse operation binding the contract event 0x57acec708e7a7b4437dd4bb623ed53257e444ac360b3fb2ddf9f157ad13e98a0.
//
// Solidity: event LeaseCreated(bytes32 indexed leaseId, address indexed spender, address indexed earner, uint256 price)
func (_LeaseAgreement *LeaseAgreementFilterer) ParseLeaseCreated(log types.Log) (*LeaseAgreementLeaseCreated, error) {
	event := new(LeaseAgreementLeaseCreated)
	if err := _LeaseAgreement.contract.UnpackLog(event, "LeaseCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LeaseAgreementLeaseExecutedIterator is returned from FilterLeaseExecuted and is used to iterate over the raw logs and unpacked data for LeaseExecuted events raised by the LeaseAgreement contract.
type LeaseAgreementLeaseExecutedIterator struct {
	Event *LeaseAgreementLeaseExecuted // Event containing the contract specifics and raw log

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
func (it *LeaseAgreementLeaseExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LeaseAgreementLeaseExecuted)
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
		it.Event = new(LeaseAgreementLeaseExecuted)
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
func (it *LeaseAgreementLeaseExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LeaseAgreementLeaseExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LeaseAgreementLeaseExecuted represents a LeaseExecuted event raised by the LeaseAgreement contract.
type LeaseAgreementLeaseExecuted struct {
	LeaseId [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterLeaseExecuted is a free log retrieval operation binding the contract event 0x49f31d6396af3c8a43db622e5ce305ec7c4a0345ba95cc5429c5c36324a228f6.
//
// Solidity: event LeaseExecuted(bytes32 indexed leaseId)
func (_LeaseAgreement *LeaseAgreementFilterer) FilterLeaseExecuted(opts *bind.FilterOpts, leaseId [][32]byte) (*LeaseAgreementLeaseExecutedIterator, error) {

	var leaseIdRule []interface{}
	for _, leaseIdItem := range leaseId {
		leaseIdRule = append(leaseIdRule, leaseIdItem)
	}

	logs, sub, err := _LeaseAgreement.contract.FilterLogs(opts, "LeaseExecuted", leaseIdRule)
	if err != nil {
		return nil, err
	}
	return &LeaseAgreementLeaseExecutedIterator{contract: _LeaseAgreement.contract, event: "LeaseExecuted", logs: logs, sub: sub}, nil
}

// WatchLeaseExecuted is a free log subscription operation binding the contract event 0x49f31d6396af3c8a43db622e5ce305ec7c4a0345ba95cc5429c5c36324a228f6.
//
// Solidity: event LeaseExecuted(bytes32 indexed leaseId)
func (_LeaseAgreement *LeaseAgreementFilterer) WatchLeaseExecuted(opts *bind.WatchOpts, sink chan<- *LeaseAgreementLeaseExecuted, leaseId [][32]byte) (event.Subscription, error) {

	var leaseIdRule []interface{}
	for _, leaseIdItem := range leaseId {
		leaseIdRule = append(leaseIdRule, leaseIdItem)
	}

	logs, sub, err := _LeaseAgreement.contract.WatchLogs(opts, "LeaseExecuted", leaseIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LeaseAgreementLeaseExecuted)
				if err := _LeaseAgreement.contract.UnpackLog(event, "LeaseExecuted", log); err != nil {
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

// ParseLeaseExecuted is a log parse operation binding the contract event 0x49f31d6396af3c8a43db622e5ce305ec7c4a0345ba95cc5429c5c36324a228f6.
//
// Solidity: event LeaseExecuted(bytes32 indexed leaseId)
func (_LeaseAgreement *LeaseAgreementFilterer) ParseLeaseExecuted(log types.Log) (*LeaseAgreementLeaseExecuted, error) {
	event := new(LeaseAgreementLeaseExecuted)
	if err := _LeaseAgreement.contract.UnpackLog(event, "LeaseExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
