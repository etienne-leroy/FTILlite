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

const CommandCumSum = "command_cumsum" // command_cumsum <hResult :: Handle→[](int64|float64|[]byte)> <hTarget :: Handle→[](int64|float64|[]byte)>

func CumSum(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])

	target, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	result, err := target.CumSum()
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, result)

	return fmt.Sprintf("array %s %s", target.TypeCode(), hResult), nil
}
