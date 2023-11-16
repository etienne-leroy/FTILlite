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

const CommandLookup = "command_lookup" // command_getitem <hResult :: Handle→[](int64|float64|[]byte)> <hTarget :: Handle→[](int64|float64|[]byte)> <hKeys :: Handle→[](int64)>

func Lookup(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])
	hKeys := variables.Handle(args[2])

	target, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	keys, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hKeys)
	if err != nil {
		return "", err
	}

	var defaultValue types.ArrayTypeVal
	if len(args) == 4 {
		defaultValue, err = variables.GetAs[types.ArrayTypeVal](s.Variables(), variables.Handle(args[3]))
		if err != nil {
			return "", err
		}
	}

	result, err := target.Lookup(keys, defaultValue)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, result)

	return fmt.Sprintf("array %s %s", result.TypeCode(), hResult), nil
}
