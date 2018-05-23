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
	"fmt"
	"time"

	"./crypto"
	"./crypto/keystore"
	"./crypto/sha3"
	corepb "./pb"
	"./util"
	"./util/byteutils"
	"github.com/gogo/protobuf/proto"
)

const (
	// TxHashByteLength invalid tx hash length(len of []byte)
	TxHashByteLength = 32
)

var (
	// TransactionMaxGasPrice max gasPrice:1 * 10 ** 12
	TransactionMaxGasPrice, _ = util.NewUint128FromString("1000000000000")

	// TransactionMaxGas max gas:50 * 10 ** 9
	TransactionMaxGas, _ = util.NewUint128FromString("50000000000")

	// TransactionGasPrice default gasPrice : 10**6
	TransactionGasPrice, _ = util.NewUint128FromInt(1000000)

	// MinGasCountPerTransaction default gas for normal transaction
	MinGasCountPerTransaction, _ = util.NewUint128FromInt(20000)

	// GasCountPerByte per byte of data attached to a transaction gas cost
	GasCountPerByte, _ = util.NewUint128FromInt(1)

	// MaxDataPayLoadLength Max data length in transaction
	MaxDataPayLoadLength = 128 * 1024
	// MaxDataBinPayloadLength Max data length in binary transaction
	MaxDataBinPayloadLength = 64

	// MaxResultLength max execution result length
	MaxResultLength = 256
)

// Transaction type is used to handle all transaction data.
type Transaction struct {
	hash      byteutils.Hash
	from      *Address
	to        *Address
	value     *util.Uint128
	nonce     uint64
	timestamp int64
	data      *corepb.Data
	chainID   uint32
	gasPrice  *util.Uint128
	gasLimit  *util.Uint128

	// Signature
	alg  keystore.Algorithm
	sign byteutils.Hash // Signature values
}

// From return from address
func (tx *Transaction) From() *Address {
	return tx.from
}

// Timestamp return timestamp
func (tx *Transaction) Timestamp() int64 {
	return tx.timestamp
}

// To return to address
func (tx *Transaction) To() *Address {
	return tx.to
}

// ChainID return chainID
func (tx *Transaction) ChainID() uint32 {
	return tx.chainID
}

// Value return tx value
func (tx *Transaction) Value() *util.Uint128 {
	return tx.value
}

// Nonce return tx nonce
func (tx *Transaction) Nonce() uint64 {
	return tx.nonce
}

// Type return tx type
func (tx *Transaction) Type() string {
	return tx.data.Type
}

// Data return tx data
func (tx *Transaction) Data() []byte {
	return tx.data.Payload
}

func (tx *Transaction) String() string {
	return fmt.Sprintf(`{"chainID":%d, "hash":"%s", "from":"%s", "to":"%s", "nonce":%d, "value":"%s", "timestamp":%d, "gasPrice":"%s", "gasLimit":"%s", "data":"%s", "type":"%s", "alg": %d, "sign":"%s"}`,
		tx.chainID,
		tx.hash.String(),
		tx.from.String(),
		tx.to.String(),
		tx.nonce,
		tx.value.String(),
		tx.timestamp,
		tx.gasPrice.String(),
		tx.gasLimit.String(),
		tx.Data(),
		tx.Type(),
	)
}

// Transactions is an alias of Transaction array.
type Transactions []*Transaction

// NewTransaction create #Transaction instance.
func NewTransaction(chainID uint32, from, to *Address, value *util.Uint128, nonce uint64, payloadType string, payload []byte, gasPrice *util.Uint128, gasLimit *util.Uint128) (*Transaction, error) {
	if gasPrice == nil || gasPrice.Cmp(util.NewUint128()) <= 0 || gasPrice.Cmp(TransactionMaxGasPrice) > 0 {
		return nil, ErrInvalidGasPrice
	}
	if gasLimit == nil || gasLimit.Cmp(util.NewUint128()) <= 0 || gasLimit.Cmp(TransactionMaxGas) > 0 {
		return nil, ErrInvalidGasLimit
	}

	if nil == from || nil == to || nil == value {
		return nil, ErrInvalidArgument
	}

	if len(payload) > MaxDataPayLoadLength {
		return nil, ErrTxDataPayLoadOutOfMaxLength
	}

	tx := &Transaction{
		from:      from,
		to:        to,
		value:     value,
		nonce:     nonce,
		timestamp: time.Now().Unix(),
		chainID:   chainID,
		data:      &corepb.Data{Type: payloadType, Payload: payload},
		gasPrice:  gasPrice,
		gasLimit:  gasLimit,
	}
	return tx, nil
}

// Hash return the hash of transaction.
func (tx *Transaction) Hash() byteutils.Hash {
	return tx.hash
}

// GasPrice returns gasPrice
func (tx *Transaction) GasPrice() *util.Uint128 {
	return tx.gasPrice
}

// GasLimit returns gasLimit
func (tx *Transaction) GasLimit() *util.Uint128 {
	return tx.gasLimit
}

// DataLen return the length of payload
func (tx *Transaction) DataLen() int {
	return len(tx.data.Payload)
}

// Signed returns sign
func (tx *Transaction) Signed() byteutils.Hash {
	return tx.sign
}

// ToProto converts domain Tx to proto Tx
func (tx *Transaction) ToProto() (proto.Message, error) {
	value, err := tx.value.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	gasPrice, err := tx.gasPrice.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	gasLimit, err := tx.gasLimit.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	return &corepb.Transaction{
		Hash:      tx.hash,
		From:      tx.from.Address,
		To:        tx.to.Address,
		Value:     value,
		Nonce:     tx.nonce,
		Timestamp: tx.timestamp,
		Data:      tx.data,
		ChainId:   tx.chainID,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Alg:       uint32(tx.alg),
		Sign:      tx.sign,
	}, nil
}

// FromProto converts proto Tx into domain Tx
func (tx *Transaction) FromProto(msg proto.Message) error {
	if msg, ok := msg.(*corepb.Transaction); ok {
		if msg != nil {
			tx.hash = msg.Hash
			from, err := AddressParseFromBytes(msg.From)
			if err != nil {
				return err
			}
			tx.from = from

			to, err := AddressParseFromBytes(msg.To)
			if err != nil {
				return err
			}
			tx.to = to

			value, err := util.NewUint128FromFixedSizeByteSlice(msg.Value)
			if err != nil {
				return err
			}
			tx.value = value

			tx.nonce = msg.Nonce
			tx.timestamp = msg.Timestamp
			tx.chainID = msg.ChainId

			if msg.Data == nil {
				return ErrInvalidTransactionData
			}
			if len(msg.Data.Payload) > MaxDataPayLoadLength {
				return ErrTxDataPayLoadOutOfMaxLength
			}
			tx.data = msg.Data

			gasPrice, err := util.NewUint128FromFixedSizeByteSlice(msg.GasPrice)
			if err != nil {
				return err
			}
			if gasPrice.Cmp(util.Uint128Zero()) <= 0 || gasPrice.Cmp(TransactionMaxGasPrice) > 0 {
				return ErrInvalidGasPrice
			}
			tx.gasPrice = gasPrice

			gasLimit, err := util.NewUint128FromFixedSizeByteSlice(msg.GasLimit)
			if err != nil {
				return err
			}
			if gasLimit.Cmp(util.Uint128Zero()) <= 0 || gasLimit.Cmp(TransactionMaxGas) > 0 {
				return ErrInvalidGasLimit
			}
			tx.gasLimit = gasLimit

			alg := keystore.Algorithm(msg.Alg)
			if err := crypto.CheckAlgorithm(alg); err != nil {
				return err
			}

			tx.alg = alg
			tx.sign = msg.Sign
			return nil
		}
		return ErrInvalidProtoToTransaction
	}
	return ErrInvalidProtoToTransaction
}

// Sign sign transaction,sign algorithm is
func (tx *Transaction) Sign(signature keystore.Signature) error {
	if signature == nil {
		return ErrNilArgument
	}
	hash, err := tx.calHash()
	if err != nil {
		return err
	}
	sign, err := signature.Sign(hash)
	if err != nil {
		return err
	}
	tx.hash = hash
	tx.alg = signature.Algorithm()
	tx.sign = sign
	return nil
}

// VerifyIntegrity return transaction verify result, including Hash and Signature.
func (tx *Transaction) VerifyIntegrity(chainID uint32) error {
	// check ChainID.
	if tx.chainID != chainID {
		return ErrInvalidChainID
	}

	// check Hash.
	wantedHash, err := tx.calHash()
	if err != nil {
		return err
	}
	if wantedHash.Equals(tx.hash) == false {
		return ErrInvalidTransactionHash
	}

	// check Signature.
	return tx.verifySign()

}

func (tx *Transaction) verifySign() error {
	signer, err := RecoverSignerFromSignature(tx.alg, tx.hash, tx.sign)
	if err != nil {
		return err
	}
	if !tx.from.Equals(signer) {
		return ErrInvalidTransactionSigner
	}
	return nil
}

// HashTransaction hash the transaction.
func (tx *Transaction) calHash() (byteutils.Hash, error) {
	hasher := sha3.New256()

	value, err := tx.value.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	data, err := proto.Marshal(tx.data)
	if err != nil {
		return nil, err
	}
	gasPrice, err := tx.gasPrice.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}
	gasLimit, err := tx.gasLimit.ToFixedSizeByteSlice()
	if err != nil {
		return nil, err
	}

	hasher.Write(tx.from.Address)
	hasher.Write(tx.to.Address)
	hasher.Write(value)
	hasher.Write(byteutils.FromUint64(tx.nonce))
	hasher.Write(byteutils.FromInt64(tx.timestamp))
	hasher.Write(data)
	hasher.Write(byteutils.FromUint32(tx.chainID))
	hasher.Write(gasPrice)
	hasher.Write(gasLimit)

	return hasher.Sum(nil), nil
}
