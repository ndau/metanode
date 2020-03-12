package state

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


import (
	"fmt"
	"reflect"

	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"
	util "github.com/ndau/noms-util"
	"github.com/pkg/errors"
)

// this code generated by github.com/ndau/generator/cmd/nomsify
// it was edited and moved to a new file in order that regenerating will cause
// compilation errors, so that humans can appropriately edit the generated code.
// search for "WARNING WARNING" to find the appropriate section.

// Adding new fields to a nomsify-able struct:
//
// Managed vars are useful for adding new fields that are marshaled to noms only after they're
// first set, so that app hashes aren't affected until the new fields are actually needed.
//
// A managed vars map is a hash map whose keys are managed variable names.
// The `managedVars map[string]struct{}` field must be manually declared in the struct.
//
// Declare new fields using the "managedVar" prefix.  e.g. `managedVarSomething SomeType`.
// GetSomething() and SetSomething() are generated for public access to the new field.
//
// Once SetSomething() is called for the first time, typically as a result of processing a new
// transaction that uses it, the managed vars map will contain "Something" as a key and the
// value of managedVarSomething will be stored in noms on the next call to MarshalNoms().
// Until then, all new managedVar fields will retain their "zero" values.

var metastateStructTemplate nt.StructTemplate

func init() {
	metastateStructTemplate = nt.MakeStructTemplate("Metastate", []string{
		"ChildState",
		"Height",
		"Stats",
		"Validators",
	})
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x Metastate) MarshalNoms(vrw nt.ValueReadWriter) (metastateValue nt.Value, err error) {
	// x.Validators (map[string]int64->*ast.MapType) is primitive: false
	// template decompose: x.Validators (map[string]int64->*ast.MapType)
	// template map: x.Validators
	validatorsKVs := make([]nt.Value, 0, len(x.Validators)*2)
	for validatorsKey, validatorsValue := range x.Validators {
		// template decompose: validatorsValue (int64->*ast.Ident)
		validatorsKVs = append(
			validatorsKVs,
			nt.String(validatorsKey),
			util.Int(validatorsValue).NomsValue(),
		)
	}

	// x.Height (uint64->*ast.Ident) is primitive: true

	// x.Stats (VoteStats->*ast.Ident) is primitive: false
	// template decompose: x.Stats (VoteStats->*ast.Ident)
	// template nomsmarshaler: x.Stats
	statsValue, err := x.Stats.MarshalNoms(vrw)
	if err != nil {
		return nil, errors.Wrap(err, "Metastate.MarshalNoms->Stats.MarshalNoms")
	}

	// x.ChildState (State->*ast.Ident) is primitive: false
	// template decompose: x.ChildState (State->*ast.Ident)
	// template nomsmarshaler: x.ChildState
	childStateValue, err := x.ChildState.MarshalNoms(vrw)
	if err != nil {
		return nil, errors.Wrap(err, "Metastate.MarshalNoms->ChildState.MarshalNoms")
	}

	values := []nt.Value{
		// x.ChildState (State)
		childStateValue,
		// x.Height (uint64)
		util.Int(x.Height).NomsValue(),
		// x.Stats (VoteStats)
		statsValue,
		// x.Validators (map[string]int64)
		nt.NewMap(vrw, validatorsKVs...),
	}

	return metastateStructTemplate.NewStruct(values), nil
}

var _ marshal.Marshaler = (*Metastate)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *Metastate) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"Metastate.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) (stop bool) {
		switch name {
		// x.Validators (map[string]int64->*ast.MapType) is primitive: false
		case "Validators":
			// template u_decompose: x.Validators (map[string]int64->*ast.MapType)
			// template u_map: x.Validators
			validatorsGMap := make(map[string]int64)
			if validatorsNMap, ok := value.(nt.Map); ok {
				validatorsNMap.Iter(func(validatorsKey, validatorsValue nt.Value) (stop bool) {
					validatorsKeyString, ok := validatorsKey.(nt.String)
					if !ok {
						err = fmt.Errorf(
							"Metastate.UnmarshalNoms expected validatorsKey to be a nt.String; found %s",
							reflect.TypeOf(validatorsKey),
						)
						return true
					}

					// template u_decompose: validatorsValue (int64->*ast.Ident)
					// template u_primitive: validatorsValue
					var validatorsValueValue util.Int
					validatorsValueValue, err = util.IntFrom(validatorsValue)
					if err != nil {
						err = errors.Wrap(err, "Metastate.UnmarshalNoms->validatorsValue")
						return
					}
					validatorsValueTyped := int64(validatorsValueValue)
					if err != nil {
						return true
					}
					validatorsGMap[string(validatorsKeyString)] = validatorsValueTyped
					return false
				})
			} else {
				err = fmt.Errorf(
					"Metastate.UnmarshalNoms expected validatorsGMap to be a nt.Map; found %s",
					reflect.TypeOf(value),
				)
			}

			x.Validators = validatorsGMap
		// x.Height (uint64->*ast.Ident) is primitive: true
		case "Height":
			// template u_decompose: x.Height (uint64->*ast.Ident)
			// template u_primitive: x.Height
			var heightValue util.Int
			heightValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "Metastate.UnmarshalNoms->Height")
				return
			}
			heightTyped := uint64(heightValue)

			x.Height = heightTyped
		// x.Stats (VoteStats->*ast.Ident) is primitive: false
		case "Stats":
			// template u_decompose: x.Stats (VoteStats->*ast.Ident)
			// template u_nomsmarshaler: x.Stats
			var statsInstance VoteStats
			err = statsInstance.UnmarshalNoms(value)
			err = errors.Wrap(err, "Metastate.UnmarshalNoms->Stats")

			x.Stats = statsInstance
		// x.ChildState (State->*ast.Ident) is primitive: false
		case "ChildState":
			// template u_decompose: x.ChildState (State->*ast.Ident)
			// template u_nomsmarshaler: x.ChildState

			// WARNING WARNING WARNING WARNING WARNING WARNING WARNING
			// this code is hand-edited: we have as a precondition that
			// x.ChildState is a non-nil instance of the State interface;
			// this is mandatory to get this unmarshaller to work.
			//
			// In general, when not dealing with interfaces, the generated
			// code will work fine, but we can't rely on that right now.
			err = x.ChildState.UnmarshalNoms(value)
			err = errors.Wrap(err, "Metastate.UnmarshalNoms->ChildState")
		}
		stop = err != nil
		return
	})
	return
}

var _ marshal.Unmarshaler = (*Metastate)(nil)
