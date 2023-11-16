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
	"errors"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const CommandSetItem = "command_setitem" // command_setitem <hTarget :: Handle→[](int64|float64|[]byte)> <hValues :: Handle→[](int64|float64|[]byte)> [ <hKeys :: Handle→[](int64)> ]

func SetItem(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hValues := variables.Handle(args[1])

	target, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	values, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hValues)
	if err != nil {
		return "", err
	}

	if target.TypeCode() != values.TypeCode() {
		return "", errors.New("values must be of the same type as the array")
	}

	var keys *types.FTIntegerArray
	if len(args) > 2 {
		hKeys := variables.Handle(args[2])

		keys, err = variables.GetAs[*types.FTIntegerArray](s.Variables(), hKeys)
		if err != nil {
			return "", err
		}

		err = target.Set(keys, values)
		if err != nil {
			return "", err
		}

		return Ack, nil
	} else {
		ys, err := values.Clone()
		if err != nil {
			return "", nil
		}
		s.Variables().Set(hTarget, ys)

		return Ack, nil
	}
}
