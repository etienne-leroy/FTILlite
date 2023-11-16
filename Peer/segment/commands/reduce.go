// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package commands

import (
	"fmt"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const (
	CommandReduceSum  = "command_reducesum"  // command_reducesum <hTarget :: Handle→[](int64|float64|[]byte)> <hValues :: Handle→[](int64|float64|[]byte)> <hKeys :: Handle→[](int64)>
	CommandReduceISum = "command_reduceisum" // command_reduceisum <hTarget :: Handle→[](int64|float64|[]byte)> <hValues :: Handle→[](int64|float64|[]byte)> <hKeys :: Handle→[](int64)>
	CommandReduceMax  = "command_reducemax"  // command_reducemax <hTarget :: Handle→[](int64|float64|[]byte)> <hValues :: Handle→[](int64|float64|[]byte)> <hKeys :: Handle→[](int64)>
	CommandReduceIMax = "command_reduceimax" // command_reduceimax <hTarget :: Handle→[](int64|float64|[]byte)> <hValues :: Handle→[](int64|float64|[]byte)> <hKeys :: Handle→[](int64)>
	CommandReduceMin  = "command_reducemin"  // command_reducemin <hTarget :: Handle→[](int64|float64|[]byte)> <hValues :: Handle→[](int64|float64|[]byte)> <hKeys :: Handle→[](int64)>
	CommandReduceIMin = "command_reduceimin" // command_reduceimin <hTarget :: Handle→[](int64|float64|[]byte)> <hValues :: Handle→[](int64|float64|[]byte)> <hKeys :: Handle→[](int64)>
)

func reduce[T types.ArrayTypeVal](s SegmentHost, args []string, f func(target T, keys *types.FTIntegerArray, values T) error) (string, error) {
	hTarget := variables.Handle(args[0])
	hValues := variables.Handle(args[1])
	hKeys := variables.Handle(args[2])

	target, err := variables.GetAs[T](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	keys, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hKeys)
	if err != nil {
		return "", err
	}

	values, err := variables.GetAs[T](s.Variables(), hValues)
	if err != nil {
		return "", err
	}

	if keys.Length() != values.Length() {
		return "", fmt.Errorf("length of keys and values do not match")
	}

	err = f(target, keys, values)
	if err != nil {
		return "", err
	}

	return Ack, nil
}

func ReduceSum(s SegmentHost, args []string) (string, error) {
	return reduce(s, args, func(target types.ArrayTypeVal, keys *types.FTIntegerArray, values types.ArrayTypeVal) error {
		return target.ReduceSum(keys, values)
	})
}

func ReduceISum(s SegmentHost, args []string) (string, error) {
	return reduce(s, args, func(target types.ArrayTypeVal, keys *types.FTIntegerArray, values types.ArrayTypeVal) error {
		return target.ReduceISum(keys, values)
	})
}

func ReduceMax(s SegmentHost, args []string) (string, error) {
	return reduce(s, args, func(target types.ArrayComparableTypeVal, keys *types.FTIntegerArray, values types.ArrayComparableTypeVal) error {
		return target.ReduceMax(keys, values)
	})
}
func ReduceIMax(s SegmentHost, args []string) (string, error) {
	return reduce(s, args, func(target types.ArrayComparableTypeVal, keys *types.FTIntegerArray, values types.ArrayComparableTypeVal) error {
		return target.ReduceIMax(keys, values)
	})
}
func ReduceMin(s SegmentHost, args []string) (string, error) {
	return reduce(s, args, func(target types.ArrayComparableTypeVal, keys *types.FTIntegerArray, values types.ArrayComparableTypeVal) error {
		return target.ReduceMin(keys, values)
	})
}
func ReduceIMin(s SegmentHost, args []string) (string, error) {
	return reduce(s, args, func(target types.ArrayComparableTypeVal, keys *types.FTIntegerArray, values types.ArrayComparableTypeVal) error {
		return target.ReduceIMin(keys, values)
	})
}
