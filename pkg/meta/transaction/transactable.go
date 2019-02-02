package metatx

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/gofrs/uuid"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// Transactable is the transaction type that the app actually cares about.
//
// Whereas Transaction is the type that tendermint cares about, with nonces
// etc designed to keep consensus running, Transactables handle application-
// specific transaction logic.
//
// Transactables are defined in terms of abci.Application. This is a
// convenience abstraction: expected behavior is to simply cast the received
// value to the type of the actual application being written.
//
// An alternative API was considered: writing a TransactableHandler interface
// with methods ValidateTransactable and ApplyTransactable, which would
// both receive the transactable. The idea was that the logic could be placed
// in whichever portion of the interface was simpler. However, that approach
// was not actually more typesafe (as it would require casting the Transactable
// to a concrete type), and would require more busywork to implement, so it
// was decided against.
type Transactable interface {
	// transactable types need to be able to marshal and unmarshal themselves
	msgp.Marshaler
	msgp.Unmarshaler

	// Validate returns nil if the Transactable is valid, or an error otherwise.
	//
	// `app` will always be an instance of your app; it is expected that the
	// first line of most implementations will be
	//
	// ```go
	// app := appInt.(*MyApp)
	// ````
	Validate(app interface{}) error

	// Apply applies this transaction to the supplied application, updating its
	// internal state as required.
	//
	// If anything but nil is returned, the internal state of the input App
	// must be unchanged.
	//
	// `app` will always be an instance of your app; it is expected that the
	// first line of most implementations will be
	//
	// ```go
	// app := appInt.(*MyApp)
	// ```
	Apply(app interface{}) error

	// SignableBytes must return a representation of all the data in the
	// transactable, except the signature. No particular ordering or schema
	// is required, only that the bytes are computed deterministically and
	// contain a complete representation of the transactable.
	//
	// For transactables with a signature field, this must _not_ include the
	// signature.
	//
	// For transactables without a signature field, this must include all data.
	// In that case, it is probably simplest to delegate this function
	// to `.MarshalMsg(nil)`.
	//
	// Unfortunately, we can't enforce the above two restrictions in code; just
	// trust that if you don't write your implementation with those semantics,
	// you will encounter weird, hard-to-debug errors eventually.
	//
	// This is designed to support two use cases:
	//
	//   1.  Most transactables must be signed. Using this method simplifies
	//       the problem of generating and validating signatures.
	//   2.  For error tracing, we desire a short and unique-ish string which
	//       can be computed for every transaction. We can get there from here.
	SignableBytes() []byte
}

// AsTransaction builds a Transaction from any Transactable
func AsTransaction(txab Transactable, idMap TxIDMap) (*Transaction, error) {
	bytes, err := txab.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}
	id, err := TxIDOf(txab, idMap)
	if err != nil {
		return nil, err
	}
	nonce, err := uuid.NewV1()
	if err != nil {
		return nil, err
	}
	return &Transaction{
		Nonce:          nonce.Bytes(),
		Transactable:   bytes,
		TransactableID: id,
	}, nil
}

func unmarshal(txab Transactable, bytes []byte) error {
	remainder, err := txab.UnmarshalMsg(bytes)
	if len(remainder) > 0 {
		return errors.New("Unmarshalling produced non-0 remainder")
	}
	return err
}

// Clone makes a shallow copy of a transactable
func Clone(original Transactable) Transactable {
	// https://stackoverflow.com/a/22948379/504550
	val := reflect.ValueOf(original)
	for val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}
	return reflect.New(val.Type()).Interface().(Transactable)
}

// AsTransactable converts a Transaction into a Transactable instance
func (tx *Transaction) AsTransactable(idMap TxIDMap) (Transactable, error) {
	instance, knownType := idMap[tx.TransactableID]
	if !knownType {
		return nil, fmt.Errorf("Unknown TransactableID: %d", tx.TransactableID)
	}

	newTxab := Clone(instance)
	err := unmarshal(newTxab, tx.Transactable)
	return newTxab, err
}

// Unmarshal constructs a Transactable from a serialized Transaction
func Unmarshal(bytes []byte, idMap TxIDMap) (Transactable, error) {
	txn := Transaction{}
	leftovers, err := txn.UnmarshalMsg(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Transaction deserialization failed")
	}
	if len(leftovers) > 0 {
		return nil, errors.New("Transaction deserialization produced leftover bytes")
	}
	return txn.AsTransactable(idMap)
}

// Marshal serializes a Transactable into a byte slice
func Marshal(txab Transactable, idMap TxIDMap) ([]byte, error) {
	tx, err := AsTransaction(txab, idMap)
	if err != nil {
		return nil, err
	}
	return tx.MarshalMsg(nil)
}

// Hash deterministically computes a probably-unique hash of the Transactable.
//
// This is intended for event tracing: it allows the Transactable to be followed
// in logs through the complete round-trip of actions.
func Hash(txab Transactable) string {
	sum := md5.Sum(txab.SignableBytes())
	return base64.RawStdEncoding.EncodeToString(sum[:])
}

// Sign the Transactable with the given private key
func Sign(txab Transactable, key signature.PrivateKey) signature.Signature {
	return key.Sign(txab.SignableBytes())
}

// Verify the Transactable's signature with the given public key
func Verify(txab Transactable, sig signature.Signature, key signature.PublicKey) bool {
	return key.Verify(txab.SignableBytes(), sig)
}
