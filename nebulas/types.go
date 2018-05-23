// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package core

import (
	"errors"
	"regexp"

	"./util"
)

// Payload Types
const (
	TxPayloadBinaryType = "binary"
	TxPayloadDeployType = "deploy"
	TxPayloadCallType   = "call"
)

// Const.
const (
	SourceTypeJavaScript = "js"
	SourceTypeTypeScript = "ts"
)

// Const
const (
	ContractAcceptFunc = "accept"
)

var (
	// PublicFuncNameChecker     in smart contract
	PublicFuncNameChecker = regexp.MustCompile("^[a-zA-Z$][A-Za-z0-9_$]*$")
)

const (
	// TxExecutionFailed failed status for transaction execute result.
	TxExecutionFailed = 0

	// TxExecutionSuccess success status for transaction execute result.
	TxExecutionSuccess = 1

	// TxExecutionPendding pendding status when transaction in transaction pool.
	TxExecutionPendding = 2
)

// Error Types
var (
	ErrInvalidChainID           = errors.New("invalid transaction chainID")
	ErrInvalidTransactionSigner = errors.New("invalid transaction signer")
	ErrInvalidTransactionHash   = errors.New("invalid transaction hash")
	ErrInvalidSignature         = errors.New("invalid transaction signature")
	ErrInvalidTxPayloadType     = errors.New("invalid transaction data payload type")
	ErrInvalidGasPrice          = errors.New("invalid gas price, should be in (0, 10^12]")
	ErrInvalidGasLimit          = errors.New("invalid gas limit, should be in (0, 5*10^10]")

	ErrTxDataPayLoadOutOfMaxLength = errors.New("data's payload is out of max data length")
	ErrNilArgument                 = errors.New("argument(s) is nil")
	ErrInvalidArgument             = errors.New("invalid argument(s)")

	ErrInsufficientBalance                = errors.New("insufficient balance")
	ErrBelowGasPrice                      = errors.New("below the gas price")
	ErrGasCntOverflow                     = errors.New("the count of gas used is overflow")
	ErrGasFeeOverflow                     = errors.New("the fee of gas used is overflow")
	ErrInvalidTransfer                    = errors.New("transfer error: overflow or insufficient balance")
	ErrGasLimitLessOrEqualToZero          = errors.New("gas limit less or equal to 0")
	ErrOutOfGasLimit                      = errors.New("out of gas limit")
	ErrTxExecutionFailed                  = errors.New("transaction execution failed")
	ErrZeroGasPrice                       = errors.New("gas price should be greater than zero")
	ErrZeroGasLimit                       = errors.New("gas limit should be greater than zero")
	ErrContractDeployFailed               = errors.New("contract deploy failed")
	ErrContractCheckFailed                = errors.New("contract check failed")
	ErrContractTransactionAddressNotEqual = errors.New("contract transaction from-address not equal to to-address")

	ErrDuplicatedTransaction = errors.New("duplicated transaction")
	ErrSmallTransactionNonce = errors.New("cannot accept a transaction with smaller nonce")
	ErrLargeTransactionNonce = errors.New("cannot accept a transaction with too bigger nonce")

	ErrInvalidAddress         = errors.New("address: invalid address")
	ErrInvalidAddressFormat   = errors.New("address: invalid address format")
	ErrInvalidAddressType     = errors.New("address: invalid address type")
	ErrInvalidAddressChecksum = errors.New("address: invalid address checksum")

	ErrInvalidProtoToBlockHeader = errors.New("protobuf message cannot be converted into BlockHeader")
	ErrInvalidProtoToTransaction = errors.New("protobuf message cannot be converted into Transaction")
	ErrInvalidTransactionData    = errors.New("invalid data in tx from Proto")

	ErrInvalidDeploySource     = errors.New("invalid source of deploy payload")
	ErrInvalidDeploySourceType = errors.New("invalid source type of deploy payload")
	ErrInvalidCallFunction     = errors.New("invalid function of call payload")
)

// Default gas count
var (
	DefaultPayloadGas, _ = util.NewUint128FromInt(1)

	// DefaultLimitsOfTotalMemorySize default limits of total memory size
	DefaultLimitsOfTotalMemorySize uint64 = 40 * 1000 * 1000
)

// // TxPayload stored in tx
// type TxPayload interface {
// 	ToBytes() ([]byte, error)
// 	BaseGasCount() *util.Uint128
// 	Execute(limitedGas *util.Uint128, tx *Transaction, block *Block, ws WorldState) (*util.Uint128, string, error)
// }

// MessageType
const (
	MessageTypeNewBlock                   = "newblock"
	MessageTypeParentBlockDownloadRequest = "dlblock"
	MessageTypeBlockDownloadResponse      = "dlreply"
	MessageTypeNewTx                      = "newtx"
)

// SyncService interface of sync service
type SyncService interface {
	Start()
	Stop()

	StartActiveSync() bool
	StopActiveSync()
	WaitingForFinish()
	IsActiveSyncing() bool
}
