package state

// this code generated by github.com/oneiro-ndev/generator/cmd/nomsify -- DO NOT EDIT

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
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
	util "github.com/oneiro-ndev/noms-util"
	"github.com/pkg/errors"
)

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

var nodeRoundStatsStructTemplate nt.StructTemplate

func init() {
	nodeRoundStatsStructTemplate = nt.MakeStructTemplate("NodeRoundStats", []string{
		"AgainstConsensus",
		"Power",
		"Voted",
	})
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x NodeRoundStats) MarshalNoms(vrw nt.ValueReadWriter) (nodeRoundStatsValue nt.Value, err error) {
	// x.Power (int64->*ast.Ident) is primitive: true

	// x.Voted (bool->*ast.Ident) is primitive: true

	// x.AgainstConsensus (bool->*ast.Ident) is primitive: true

	values := []nt.Value{
		// x.AgainstConsensus (bool)
		nt.Bool(x.AgainstConsensus),
		// x.Power (int64)
		util.Int(x.Power).NomsValue(),
		// x.Voted (bool)
		nt.Bool(x.Voted),
	}

	return nodeRoundStatsStructTemplate.NewStruct(values), nil
}

var _ marshal.Marshaler = (*NodeRoundStats)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *NodeRoundStats) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"NodeRoundStats.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) (stop bool) {
		switch name {
		// x.Power (int64->*ast.Ident) is primitive: true
		case "Power":
			// template u_decompose: x.Power (int64->*ast.Ident)
			// template u_primitive: x.Power
			var powerValue util.Int
			powerValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "NodeRoundStats.UnmarshalNoms->Power")
				return
			}
			powerTyped := int64(powerValue)

			x.Power = powerTyped
		// x.Voted (bool->*ast.Ident) is primitive: true
		case "Voted":
			// template u_decompose: x.Voted (bool->*ast.Ident)
			// template u_primitive: x.Voted
			votedValue, ok := value.(nt.Bool)
			if !ok {
				err = fmt.Errorf(
					"NodeRoundStats.UnmarshalNoms expected value to be a nt.Bool; found %s",
					reflect.TypeOf(value),
				)
			}
			votedTyped := bool(votedValue)

			x.Voted = votedTyped
		// x.AgainstConsensus (bool->*ast.Ident) is primitive: true
		case "AgainstConsensus":
			// template u_decompose: x.AgainstConsensus (bool->*ast.Ident)
			// template u_primitive: x.AgainstConsensus
			againstConsensusValue, ok := value.(nt.Bool)
			if !ok {
				err = fmt.Errorf(
					"NodeRoundStats.UnmarshalNoms expected value to be a nt.Bool; found %s",
					reflect.TypeOf(value),
				)
			}
			againstConsensusTyped := bool(againstConsensusValue)

			x.AgainstConsensus = againstConsensusTyped
		}
		stop = err != nil
		return
	})
	return
}

var _ marshal.Unmarshaler = (*NodeRoundStats)(nil)

var roundStatsStructTemplate nt.StructTemplate

func init() {
	roundStatsStructTemplate = nt.MakeStructTemplate("RoundStats", []string{
		"Height",
		"Validators",
	})
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x RoundStats) MarshalNoms(vrw nt.ValueReadWriter) (roundStatsValue nt.Value, err error) {
	// x.Height (uint64->*ast.Ident) is primitive: true

	// x.Validators (map[string]NodeRoundStats->*ast.MapType) is primitive: false
	// template decompose: x.Validators (map[string]NodeRoundStats->*ast.MapType)
	// template map: x.Validators
	validatorsKVs := make([]nt.Value, 0, len(x.Validators)*2)
	for validatorsKey, validatorsValue := range x.Validators {
		// template decompose: validatorsValue (NodeRoundStats->*ast.Ident)
		// template nomsmarshaler: validatorsValue
		validatorsValueValue, err := validatorsValue.MarshalNoms(vrw)
		if err != nil {
			return nil, errors.Wrap(err, "RoundStats.MarshalNoms->validatorsValue.MarshalNoms")
		}
		validatorsKVs = append(
			validatorsKVs,
			nt.String(validatorsKey),
			validatorsValueValue,
		)
	}

	values := []nt.Value{
		// x.Height (uint64)
		util.Int(x.Height).NomsValue(),
		// x.Validators (map[string]NodeRoundStats)
		nt.NewMap(vrw, validatorsKVs...),
	}

	return roundStatsStructTemplate.NewStruct(values), nil
}

var _ marshal.Marshaler = (*RoundStats)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *RoundStats) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"RoundStats.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) (stop bool) {
		switch name {
		// x.Height (uint64->*ast.Ident) is primitive: true
		case "Height":
			// template u_decompose: x.Height (uint64->*ast.Ident)
			// template u_primitive: x.Height
			var heightValue util.Int
			heightValue, err = util.IntFrom(value)
			if err != nil {
				err = errors.Wrap(err, "RoundStats.UnmarshalNoms->Height")
				return
			}
			heightTyped := uint64(heightValue)

			x.Height = heightTyped
		// x.Validators (map[string]NodeRoundStats->*ast.MapType) is primitive: false
		case "Validators":
			// template u_decompose: x.Validators (map[string]NodeRoundStats->*ast.MapType)
			// template u_map: x.Validators
			validatorsGMap := make(map[string]NodeRoundStats)
			if validatorsNMap, ok := value.(nt.Map); ok {
				validatorsNMap.Iter(func(validatorsKey, validatorsValue nt.Value) (stop bool) {
					validatorsKeyString, ok := validatorsKey.(nt.String)
					if !ok {
						err = fmt.Errorf(
							"RoundStats.UnmarshalNoms expected validatorsKey to be a nt.String; found %s",
							reflect.TypeOf(validatorsKey),
						)
						return true
					}

					// template u_decompose: validatorsValue (NodeRoundStats->*ast.Ident)
					// template u_nomsmarshaler: validatorsValue
					var validatorsValueInstance NodeRoundStats
					err = validatorsValueInstance.UnmarshalNoms(validatorsValue)
					err = errors.Wrap(err, "RoundStats.UnmarshalNoms->validatorsValue")
					if err != nil {
						return true
					}
					validatorsGMap[string(validatorsKeyString)] = validatorsValueInstance
					return false
				})
			} else {
				err = fmt.Errorf(
					"RoundStats.UnmarshalNoms expected validatorsGMap to be a nt.Map; found %s",
					reflect.TypeOf(value),
				)
			}

			x.Validators = validatorsGMap
		}
		stop = err != nil
		return
	})
	return
}

var _ marshal.Unmarshaler = (*RoundStats)(nil)

var voteStatsStructTemplate nt.StructTemplate

func init() {
	voteStatsStructTemplate = nt.MakeStructTemplate("VoteStats", []string{
		"History",
	})
}

// MarshalNoms implements noms/go/marshal.Marshaler
func (x VoteStats) MarshalNoms(vrw nt.ValueReadWriter) (voteStatsValue nt.Value, err error) {
	// x.History ([]RoundStats->*ast.ArrayType) is primitive: false
	// template decompose: x.History ([]RoundStats->*ast.ArrayType)
	// template slice: x.History
	historyItems := make([]nt.Value, 0, len(x.History))
	for _, historyItem := range x.History {
		// template decompose: historyItem (RoundStats->*ast.Ident)
		// template nomsmarshaler: historyItem
		historyItemValue, err := historyItem.MarshalNoms(vrw)
		if err != nil {
			return nil, errors.Wrap(err, "VoteStats.MarshalNoms->historyItem.MarshalNoms")
		}
		historyItems = append(
			historyItems,
			historyItemValue,
		)
	}

	values := []nt.Value{
		// x.History ([]RoundStats)
		nt.NewList(vrw, historyItems...),
	}

	return voteStatsStructTemplate.NewStruct(values), nil
}

var _ marshal.Marshaler = (*VoteStats)(nil)

// UnmarshalNoms implements noms/go/marshal.Unmarshaler
//
// This method makes no attempt to zeroize the provided struct; it simply
// overwrites fields as they are found.
func (x *VoteStats) UnmarshalNoms(value nt.Value) (err error) {
	vs, ok := value.(nt.Struct)
	if !ok {
		return fmt.Errorf(
			"VoteStats.UnmarshalNoms expected a nt.Value; found %s",
			reflect.TypeOf(value),
		)
	}

	// noms Struct.MaybeGet isn't efficient: it iterates over all fields of
	// the struct until it finds one whose name happens to match the one sought.
	// It's better to iterate once over the struct and set the fields of the
	// target struct in arbitrary order.
	vs.IterFields(func(name string, value nt.Value) (stop bool) {
		switch name {
		// x.History ([]RoundStats->*ast.ArrayType) is primitive: false
		case "History":
			// template u_decompose: x.History ([]RoundStats->*ast.ArrayType)
			// template u_slice: x.History
			var historySlice []RoundStats
			if historyList, ok := value.(nt.List); ok {
				historySlice = make([]RoundStats, 0, historyList.Len())
				historyList.Iter(func(historyItem nt.Value, idx uint64) (stop bool) {

					// template u_decompose: historyItem (RoundStats->*ast.Ident)
					// template u_nomsmarshaler: historyItem
					var historyItemInstance RoundStats
					err = historyItemInstance.UnmarshalNoms(historyItem)
					err = errors.Wrap(err, "VoteStats.UnmarshalNoms->historyItem")
					if err != nil {
						return true
					}
					historySlice = append(historySlice, historyItemInstance)
					return false
				})
			} else {
				err = fmt.Errorf(
					"VoteStats.UnmarshalNoms expected value to be a nt.List; found %s",
					reflect.TypeOf(value),
				)
			}

			x.History = historySlice
		}
		stop = err != nil
		return
	})
	return
}

var _ marshal.Unmarshaler = (*VoteStats)(nil)
