package metatx

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	abci "github.com/tendermint/abci/types"
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

	// IsValid returns nil if the Transactable is valid, or an error otherwise.
	IsValid(app abci.Application) error
	// Apply applies this transaction to the supplied application, updating its
	// internal state as required.
	//
	// If anything but nil is returned, the internal state of the input App
	// must be unchanged.
	Apply(app abci.Application) error
}

// AsTransaction builds a Transaction from any Transactable
func AsTransaction(txab Transactable, idMap TxIDMap) (*Transaction, error) {
	bytes, err := txab.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}
	id, err := txIDOf(txab, idMap)
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

// shallow copy an interface from an example struct
// https://stackoverflow.com/a/22948379/504550
func cloneTxab(original Transactable) Transactable {
	val := reflect.ValueOf(original)
	if val.Kind() == reflect.Ptr {
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

	newTxab := cloneTxab(instance)
	err := unmarshal(newTxab, tx.Transactable)
	return newTxab, err
}

// TransactableFromBytes constructs a Transactable from a serialized Transaction
func TransactableFromBytes(bytes []byte, idMap TxIDMap) (Transactable, error) {
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

// TransactableToBytes serializes a Transactable into a byte slice
func TransactableToBytes(txab Transactable, idMap TxIDMap) ([]byte, error) {
	tx, err := AsTransaction(txab, idMap)
	if err != nil {
		return nil, err
	}
	return tx.MarshalMsg(nil)
}
