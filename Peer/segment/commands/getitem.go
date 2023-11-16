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

const CommandGetItem = "command_getitem" // command_getitem <hResult :: Handle→[](int64|float64|[]byte)> <hTarget :: Handle→[](int64|float64|[]byte)> [ <hKeys :: Handle→[](int64)> ]

func GetItem(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])

	target, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	var result types.ArrayTypeVal
	if len(args) > 2 {
		hKeys := variables.Handle(args[2])

		keys, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hKeys)
		if err != nil {
			return "", err
		}

		result, err = target.Get(keys, nil)
		if err != nil {
			return "", err
		}

	} else {
		result, err = target.Get(nil, nil)
		if err != nil {
			return "", err
		}
	}

	s.Variables().Set(hResult, result)

	return fmt.Sprintf("array %s %s", result.TypeCode(), hResult), nil
}
